package proxmox

import (
	"context"
	"errors"
	"fmt"
)

// Notifications lists the resource-type directory under /cluster/notifications.
func (cl *Cluster) Notifications(ctx context.Context) (entries ClusterNotificationIndex, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications", &entries)
	return
}

// NotificationMatcherFields returns the known metadata field names usable in
// matcher "match-field" rules.
func (cl *Cluster) NotificationMatcherFields(ctx context.Context) (fields []*ClusterNotificationMatcherField, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications/matcher-fields", &fields)
	return
}

// NotificationMatcherFieldValues returns the known (field, value) pairs for
// matcher exact-match rules.
func (cl *Cluster) NotificationMatcherFieldValues(ctx context.Context) (values []*ClusterNotificationMatcherFieldValue, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications/matcher-field-values", &values)
	return
}

// --- targets ----------------------------------------------------------------

// NotificationTargets lists all notification targets (flattened view across
// every endpoint plugin type).
func (cl *Cluster) NotificationTargets(ctx context.Context) (targets []*ClusterNotificationTarget, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications/targets", &targets)
	return
}

// TestNotificationTarget triggers PVE to send a test notification through the
// named target. Returns immediately; PVE does not surface a UPID for this op.
func (cl *Cluster) TestNotificationTarget(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("notification target name can not be empty")
	}
	return cl.client.Post(ctx, fmt.Sprintf("/cluster/notifications/targets/%s/test", name), nil, nil)
}

// --- matchers ---------------------------------------------------------------

// NotificationMatchers lists configured matchers.
func (cl *Cluster) NotificationMatchers(ctx context.Context) (matchers []*ClusterNotificationMatcher, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications/matchers", &matchers)
	return
}

// NotificationMatcher reads a single matcher by name.
func (cl *Cluster) NotificationMatcher(ctx context.Context, name string) (m *ClusterNotificationMatcher, err error) {
	if name == "" {
		err = errors.New("notification matcher name can not be empty")
		return
	}
	m = &ClusterNotificationMatcher{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/notifications/matchers/%s", name), m); err != nil {
		return
	}
	if m.Name == "" {
		m.Name = name
	}
	return
}

// NewNotificationMatcher creates a matcher. opts.Name is required.
func (cl *Cluster) NewNotificationMatcher(ctx context.Context, opts *ClusterNotificationMatcherOptions) error {
	if opts == nil || opts.Name == "" {
		return errors.New("notification matcher name can not be empty")
	}
	return cl.client.Post(ctx, "/cluster/notifications/matchers", opts, nil)
}

// UpdateNotificationMatcher mutates an existing matcher.
func (cl *Cluster) UpdateNotificationMatcher(ctx context.Context, name string, opts *ClusterNotificationMatcherOptions) error {
	if name == "" {
		return errors.New("notification matcher name can not be empty")
	}
	if opts == nil {
		opts = &ClusterNotificationMatcherOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/notifications/matchers/%s", name), opts, nil)
}

// DeleteNotificationMatcher removes a matcher.
func (cl *Cluster) DeleteNotificationMatcher(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("notification matcher name can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/notifications/matchers/%s", name), nil)
}

// --- gotify endpoints -------------------------------------------------------

// NotificationGotifyEndpoints lists configured Gotify endpoints.
func (cl *Cluster) NotificationGotifyEndpoints(ctx context.Context) (endpoints []*ClusterNotificationGotifyEndpoint, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications/endpoints/gotify", &endpoints)
	return
}

// NotificationGotifyEndpoint reads a single Gotify endpoint.
func (cl *Cluster) NotificationGotifyEndpoint(ctx context.Context, name string) (e *ClusterNotificationGotifyEndpoint, err error) {
	if name == "" {
		err = errors.New("gotify endpoint name can not be empty")
		return
	}
	e = &ClusterNotificationGotifyEndpoint{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/notifications/endpoints/gotify/%s", name), e); err != nil {
		return
	}
	if e.Name == "" {
		e.Name = name
	}
	return
}

// NewNotificationGotifyEndpoint creates a Gotify endpoint. opts.Name, .Server,
// and .Token are required by PVE on create.
func (cl *Cluster) NewNotificationGotifyEndpoint(ctx context.Context, opts *ClusterNotificationGotifyOptions) error {
	if opts == nil || opts.Name == "" {
		return errors.New("gotify endpoint name can not be empty")
	}
	return cl.client.Post(ctx, "/cluster/notifications/endpoints/gotify", opts, nil)
}

// UpdateNotificationGotifyEndpoint mutates an existing Gotify endpoint.
func (cl *Cluster) UpdateNotificationGotifyEndpoint(ctx context.Context, name string, opts *ClusterNotificationGotifyOptions) error {
	if name == "" {
		return errors.New("gotify endpoint name can not be empty")
	}
	if opts == nil {
		opts = &ClusterNotificationGotifyOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/notifications/endpoints/gotify/%s", name), opts, nil)
}

// DeleteNotificationGotifyEndpoint removes a Gotify endpoint.
func (cl *Cluster) DeleteNotificationGotifyEndpoint(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("gotify endpoint name can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/notifications/endpoints/gotify/%s", name), nil)
}

// --- sendmail endpoints -----------------------------------------------------

// NotificationSendmailEndpoints lists configured sendmail endpoints.
func (cl *Cluster) NotificationSendmailEndpoints(ctx context.Context) (endpoints []*ClusterNotificationSendmailEndpoint, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications/endpoints/sendmail", &endpoints)
	return
}

// NotificationSendmailEndpoint reads a single sendmail endpoint.
func (cl *Cluster) NotificationSendmailEndpoint(ctx context.Context, name string) (e *ClusterNotificationSendmailEndpoint, err error) {
	if name == "" {
		err = errors.New("sendmail endpoint name can not be empty")
		return
	}
	e = &ClusterNotificationSendmailEndpoint{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/notifications/endpoints/sendmail/%s", name), e); err != nil {
		return
	}
	if e.Name == "" {
		e.Name = name
	}
	return
}

// NewNotificationSendmailEndpoint creates a sendmail endpoint.
func (cl *Cluster) NewNotificationSendmailEndpoint(ctx context.Context, opts *ClusterNotificationSendmailOptions) error {
	if opts == nil || opts.Name == "" {
		return errors.New("sendmail endpoint name can not be empty")
	}
	return cl.client.Post(ctx, "/cluster/notifications/endpoints/sendmail", opts, nil)
}

// UpdateNotificationSendmailEndpoint mutates an existing sendmail endpoint.
func (cl *Cluster) UpdateNotificationSendmailEndpoint(ctx context.Context, name string, opts *ClusterNotificationSendmailOptions) error {
	if name == "" {
		return errors.New("sendmail endpoint name can not be empty")
	}
	if opts == nil {
		opts = &ClusterNotificationSendmailOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/notifications/endpoints/sendmail/%s", name), opts, nil)
}

// DeleteNotificationSendmailEndpoint removes a sendmail endpoint.
func (cl *Cluster) DeleteNotificationSendmailEndpoint(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("sendmail endpoint name can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/notifications/endpoints/sendmail/%s", name), nil)
}

// --- smtp endpoints ---------------------------------------------------------

// NotificationSMTPEndpoints lists configured SMTP endpoints.
func (cl *Cluster) NotificationSMTPEndpoints(ctx context.Context) (endpoints []*ClusterNotificationSMTPEndpoint, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications/endpoints/smtp", &endpoints)
	return
}

// NotificationSMTPEndpoint reads a single SMTP endpoint.
func (cl *Cluster) NotificationSMTPEndpoint(ctx context.Context, name string) (e *ClusterNotificationSMTPEndpoint, err error) {
	if name == "" {
		err = errors.New("smtp endpoint name can not be empty")
		return
	}
	e = &ClusterNotificationSMTPEndpoint{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/notifications/endpoints/smtp/%s", name), e); err != nil {
		return
	}
	if e.Name == "" {
		e.Name = name
	}
	return
}

// NewNotificationSMTPEndpoint creates an SMTP endpoint.
func (cl *Cluster) NewNotificationSMTPEndpoint(ctx context.Context, opts *ClusterNotificationSMTPOptions) error {
	if opts == nil || opts.Name == "" {
		return errors.New("smtp endpoint name can not be empty")
	}
	return cl.client.Post(ctx, "/cluster/notifications/endpoints/smtp", opts, nil)
}

// UpdateNotificationSMTPEndpoint mutates an existing SMTP endpoint.
func (cl *Cluster) UpdateNotificationSMTPEndpoint(ctx context.Context, name string, opts *ClusterNotificationSMTPOptions) error {
	if name == "" {
		return errors.New("smtp endpoint name can not be empty")
	}
	if opts == nil {
		opts = &ClusterNotificationSMTPOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/notifications/endpoints/smtp/%s", name), opts, nil)
}

// DeleteNotificationSMTPEndpoint removes an SMTP endpoint.
func (cl *Cluster) DeleteNotificationSMTPEndpoint(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("smtp endpoint name can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/notifications/endpoints/smtp/%s", name), nil)
}

// --- webhook endpoints ------------------------------------------------------

// NotificationWebhookEndpoints lists configured webhook endpoints.
func (cl *Cluster) NotificationWebhookEndpoints(ctx context.Context) (endpoints []*ClusterNotificationWebhookEndpoint, err error) {
	err = cl.client.Get(ctx, "/cluster/notifications/endpoints/webhook", &endpoints)
	return
}

// NotificationWebhookEndpoint reads a single webhook endpoint.
func (cl *Cluster) NotificationWebhookEndpoint(ctx context.Context, name string) (e *ClusterNotificationWebhookEndpoint, err error) {
	if name == "" {
		err = errors.New("webhook endpoint name can not be empty")
		return
	}
	e = &ClusterNotificationWebhookEndpoint{}
	if err = cl.client.Get(ctx, fmt.Sprintf("/cluster/notifications/endpoints/webhook/%s", name), e); err != nil {
		return
	}
	if e.Name == "" {
		e.Name = name
	}
	return
}

// NewNotificationWebhookEndpoint creates a webhook endpoint.
func (cl *Cluster) NewNotificationWebhookEndpoint(ctx context.Context, opts *ClusterNotificationWebhookOptions) error {
	if opts == nil || opts.Name == "" {
		return errors.New("webhook endpoint name can not be empty")
	}
	return cl.client.Post(ctx, "/cluster/notifications/endpoints/webhook", opts, nil)
}

// UpdateNotificationWebhookEndpoint mutates an existing webhook endpoint.
func (cl *Cluster) UpdateNotificationWebhookEndpoint(ctx context.Context, name string, opts *ClusterNotificationWebhookOptions) error {
	if name == "" {
		return errors.New("webhook endpoint name can not be empty")
	}
	if opts == nil {
		opts = &ClusterNotificationWebhookOptions{}
	}
	return cl.client.Put(ctx, fmt.Sprintf("/cluster/notifications/endpoints/webhook/%s", name), opts, nil)
}

// DeleteNotificationWebhookEndpoint removes a webhook endpoint.
func (cl *Cluster) DeleteNotificationWebhookEndpoint(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("webhook endpoint name can not be empty")
	}
	return cl.client.Delete(ctx, fmt.Sprintf("/cluster/notifications/endpoints/webhook/%s", name), nil)
}
