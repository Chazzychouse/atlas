package mail

import (
	"fmt"
	"sync"

	"github.com/chazzychouse/atlas/internal/config"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

const PageSize = 50

// IMAPClient wraps go-imap/v2 for typed mailbox operations.
type IMAPClient struct {
	cfg    *config.Config
	client *imapclient.Client
	mu     sync.Mutex
}

// NewIMAPClient creates a new IMAP client (does not connect yet).
func NewIMAPClient(cfg *config.Config) *IMAPClient {
	return &IMAPClient{cfg: cfg}
}

// Connect establishes a TLS connection and authenticates.
func (c *IMAPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Close()
	}

	addr := fmt.Sprintf("%s:%d", c.cfg.IMAPHost, c.cfg.IMAPPort)
	client, err := imapclient.DialTLS(addr, nil)
	if err != nil {
		return fmt.Errorf("IMAP dial: %w", err)
	}

	if err := client.Login(c.cfg.IMAPUser, c.cfg.IMAPPass).Wait(); err != nil {
		client.Close()
		return fmt.Errorf("IMAP login: %w", err)
	}

	c.client = client
	return nil
}

// ensureConnected reconnects if the client is nil.
func (c *IMAPClient) ensureConnected() error {
	if c.client == nil {
		return c.Connect()
	}
	return nil
}

// Close closes the IMAP connection.
func (c *IMAPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		return err
	}
	return nil
}

// ListFolders returns all mailbox folders.
func (c *IMAPClient) ListFolders() ([]Folder, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	listCmd := c.client.List("", "*", nil)
	mailboxes, err := listCmd.Collect()
	if err != nil {
		return nil, fmt.Errorf("listing folders: %w", err)
	}

	var folders []Folder
	for _, mb := range mailboxes {
		var attrs []string
		for _, a := range mb.Attrs {
			attrs = append(attrs, string(a))
		}
		folders = append(folders, Folder{
			Name:       mb.Mailbox,
			Attributes: attrs,
		})
	}
	return folders, nil
}

// SelectFolder selects a mailbox and returns the number of messages.
func (c *IMAPClient) SelectFolder(name string) (uint32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return 0, err
	}

	data, err := c.client.Select(name, nil).Wait()
	if err != nil {
		return 0, fmt.Errorf("selecting folder %q: %w", name, err)
	}

	return data.NumMessages, nil
}

// FetchEnvelopes fetches envelope data for messages in the given sequence range.
// start and end are 1-based sequence numbers. Use 0 for end to mean "*".
func (c *IMAPClient) FetchEnvelopes(start, end uint32) ([]Envelope, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	var seqSet imap.SeqSet
	if end == 0 {
		seqSet.AddRange(start, 0)
	} else {
		seqSet.AddRange(start, end)
	}

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		Flags:    true,
		UID:      true,
	}

	fetchCmd := c.client.Fetch(seqSet, fetchOptions)
	var envelopes []Envelope

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		buf, err := msg.Collect()
		if err != nil {
			continue
		}

		env := Envelope{
			UID:  uint32(buf.UID),
			Seen: containsFlag(buf.Flags, imap.FlagSeen),
		}

		if buf.Envelope != nil {
			env.Subject = buf.Envelope.Subject
			env.Date = buf.Envelope.Date
			if len(buf.Envelope.From) > 0 {
				env.From = formatIMAPAddress(buf.Envelope.From[0])
			}
			for _, addr := range buf.Envelope.To {
				env.To = append(env.To, formatIMAPAddress(addr))
			}
			for _, addr := range buf.Envelope.Cc {
				env.Cc = append(env.Cc, formatIMAPAddress(addr))
			}
		}

		envelopes = append(envelopes, env)
	}

	if err := fetchCmd.Close(); err != nil {
		return nil, fmt.Errorf("fetching envelopes: %w", err)
	}

	return envelopes, nil
}

// FetchMessage fetches the full message (envelope + body) by UID.
func (c *IMAPClient) FetchMessage(uid uint32) (*Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	var uidSet imap.UIDSet
	uidSet.AddNum(imap.UID(uid))

	bodySection := &imap.FetchItemBodySection{}
	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		Flags:       true,
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{bodySection},
	}

	fetchCmd := c.client.Fetch(uidSet, fetchOptions)

	msg := fetchCmd.Next()
	if msg == nil {
		fetchCmd.Close()
		return nil, fmt.Errorf("message UID %d not found", uid)
	}

	buf, err := msg.Collect()
	if err != nil {
		fetchCmd.Close()
		return nil, fmt.Errorf("collecting message: %w", err)
	}

	if err := fetchCmd.Close(); err != nil {
		return nil, fmt.Errorf("fetching message: %w", err)
	}

	result := &Message{
		Envelope: Envelope{
			UID:  uint32(buf.UID),
			Seen: containsFlag(buf.Flags, imap.FlagSeen),
		},
	}

	if buf.Envelope != nil {
		result.Subject = buf.Envelope.Subject
		result.Date = buf.Envelope.Date
		if len(buf.Envelope.From) > 0 {
			result.From = formatIMAPAddress(buf.Envelope.From[0])
		}
		for _, addr := range buf.Envelope.To {
			result.To = append(result.To, formatIMAPAddress(addr))
		}
		for _, addr := range buf.Envelope.Cc {
			result.Cc = append(result.Cc, formatIMAPAddress(addr))
		}
	}

	// Extract body from the fetched section
	for _, section := range buf.BodySection {
		body, err := ParsePlainText(section.Bytes)
		if err == nil && body != "" {
			result.Body = body
			break
		}
		// Fallback: use raw content if MIME parsing fails
		if result.Body == "" {
			result.Body = string(section.Bytes)
		}
	}

	return result, nil
}

// DeleteMessage marks a message as deleted by UID and expunges.
func (c *IMAPClient) DeleteMessage(uid uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	var uidSet imap.UIDSet
	uidSet.AddNum(imap.UID(uid))

	storeFlags := &imap.StoreFlags{
		Op:     imap.StoreFlagsAdd,
		Silent: true,
		Flags:  []imap.Flag{imap.FlagDeleted},
	}

	storeCmd := c.client.Store(uidSet, storeFlags, nil)
	if err := storeCmd.Close(); err != nil {
		return fmt.Errorf("flagging deleted: %w", err)
	}

	expungeCmd := c.client.Expunge()
	if err := expungeCmd.Close(); err != nil {
		return fmt.Errorf("expunging: %w", err)
	}

	return nil
}

// DeleteMessages marks multiple messages as deleted by UID and expunges.
func (c *IMAPClient) DeleteMessages(uids []uint32) error {
	if len(uids) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	var uidSet imap.UIDSet
	for _, uid := range uids {
		uidSet.AddNum(imap.UID(uid))
	}

	storeFlags := &imap.StoreFlags{
		Op:     imap.StoreFlagsAdd,
		Silent: true,
		Flags:  []imap.Flag{imap.FlagDeleted},
	}

	storeCmd := c.client.Store(uidSet, storeFlags, nil)
	if err := storeCmd.Close(); err != nil {
		return fmt.Errorf("flagging deleted: %w", err)
	}

	expungeCmd := c.client.Expunge()
	if err := expungeCmd.Close(); err != nil {
		return fmt.Errorf("expunging: %w", err)
	}

	return nil
}

// SetSeenBulk marks/unmarks multiple messages as read by UID.
func (c *IMAPClient) SetSeenBulk(uids []uint32, seen bool) error {
	if len(uids) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	var uidSet imap.UIDSet
	for _, uid := range uids {
		uidSet.AddNum(imap.UID(uid))
	}

	op := imap.StoreFlagsAdd
	if !seen {
		op = imap.StoreFlagsDel
	}

	storeFlags := &imap.StoreFlags{
		Op:     op,
		Silent: true,
		Flags:  []imap.Flag{imap.FlagSeen},
	}

	storeCmd := c.client.Store(uidSet, storeFlags, nil)
	if err := storeCmd.Close(); err != nil {
		return fmt.Errorf("setting seen flag: %w", err)
	}

	return nil
}

// MoveMessage moves a message to another folder by UID.
func (c *IMAPClient) MoveMessage(uid uint32, destFolder string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	var uidSet imap.UIDSet
	uidSet.AddNum(imap.UID(uid))

	if _, err := c.client.Move(uidSet, destFolder).Wait(); err != nil {
		return fmt.Errorf("moving message: %w", err)
	}

	return nil
}

// SetSeen marks/unmarks a message as read by UID.
func (c *IMAPClient) SetSeen(uid uint32, seen bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	var uidSet imap.UIDSet
	uidSet.AddNum(imap.UID(uid))

	op := imap.StoreFlagsAdd
	if !seen {
		op = imap.StoreFlagsDel
	}

	storeFlags := &imap.StoreFlags{
		Op:     op,
		Silent: true,
		Flags:  []imap.Flag{imap.FlagSeen},
	}

	storeCmd := c.client.Store(uidSet, storeFlags, nil)
	if err := storeCmd.Close(); err != nil {
		return fmt.Errorf("setting seen flag: %w", err)
	}

	return nil
}

func containsFlag(flags []imap.Flag, target imap.Flag) bool {
	for _, f := range flags {
		if f == target {
			return true
		}
	}
	return false
}

func formatIMAPAddress(addr imap.Address) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s>", addr.Name, addr.Addr())
	}
	return addr.Addr()
}
