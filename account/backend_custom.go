package account

// CustomTopicKind is the unique identifier for custom/third-party backends.
// Use this constant when creating CommAddresses for custom endpoints.
const CustomTopicKind = "custom"

// init registers the custom backend at package initialization.
func init() {
	RegisterBackend(&customBackend{})
}

// customBackend implements the Backend interface for custom/third-party endpoints.
// This provides a generic fallback for backends that don't have specific implementations.
type customBackend struct{}

// Kind returns the unique identifier for the custom backend.
func (c *customBackend) Kind() string {
	return CustomTopicKind
}

// ValidateLocator checks if a locator is valid for custom backends.
// Custom backends only require a non-empty locator.
func (c *customBackend) ValidateLocator(locator string) error {
	if locator == "" {
		return errInvalidAddress("customBackend.ValidateLocator", "empty custom locator", nil)
	}
	return nil
}

// ParseLocator validates and returns the locator unchanged.
// Custom locators are accepted as-is without normalization.
func (c *customBackend) ParseLocator(locator string) (string, error) {
	if err := c.ValidateLocator(locator); err != nil {
		return "", err
	}
	return locator, nil
}

// Metadata returns descriptive information about the custom backend.
func (c *customBackend) Metadata() BackendMetadata {
	return BackendMetadata{
		DisplayName:    "Custom Backend",
		Description:    "Generic custom communication endpoint for third-party integrations",
		LocatorFormat:  "implementation-defined",
		LocatorExample: "my-custom-endpoint",
		RequiresConfig: false,
		Version:        "1.0.0",
		Properties:     map[string]string{},
	}
}

// Technology returns the TopicTechnology for the custom backend.
func (c *customBackend) Technology() TopicTechnology {
	return TopicTechnologyCustom
}

// NewCustomTopicAddress creates a CommAddress with the custom backend.
// This is useful for third-party or custom communication endpoints.
func NewCustomTopicAddress(locator string) (CommAddress, error) {
	return NewCommAddress(CustomTopicKind, locator)
}
