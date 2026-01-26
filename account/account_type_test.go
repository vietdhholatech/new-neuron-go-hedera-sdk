package account

import (
	"errors"
	"testing"
)

// =============================================================================
// AccountType.String Tests
// =============================================================================

func TestAccountType_String(t *testing.T) {
	tests := []struct {
		accountType AccountType
		expected    string
	}{
		{AccountTypeUnspecified, "Unspecified"},
		{AccountTypeParent, "Parent"},
		{AccountTypeChild, "Child"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.accountType.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}

	t.Run("unknown value returns Unknown(N)", func(t *testing.T) {
		unknownType := AccountType(999)
		got := unknownType.String()
		expected := "Unknown(999)"
		if got != expected {
			t.Errorf("String() = %v, want %v", got, expected)
		}
	})

	t.Run("negative value returns Unknown(N)", func(t *testing.T) {
		negativeType := AccountType(-1)
		got := negativeType.String()
		expected := "Unknown(-1)"
		if got != expected {
			t.Errorf("String() = %v, want %v", got, expected)
		}
	})
}

// =============================================================================
// AccountType.IsValid Tests
// =============================================================================

func TestAccountType_IsValid(t *testing.T) {
	t.Run("Parent is valid", func(t *testing.T) {
		if !AccountTypeParent.IsValid() {
			t.Error("Parent should be valid")
		}
	})

	t.Run("Child is valid", func(t *testing.T) {
		if !AccountTypeChild.IsValid() {
			t.Error("Child should be valid")
		}
	})

	t.Run("Unspecified is invalid", func(t *testing.T) {
		if AccountTypeUnspecified.IsValid() {
			t.Error("Unspecified should be invalid")
		}
	})

	t.Run("zero value is invalid", func(t *testing.T) {
		var zeroType AccountType
		if zeroType.IsValid() {
			t.Error("zero value should be invalid")
		}
	})

	t.Run("negative value is invalid", func(t *testing.T) {
		negativeType := AccountType(-1)
		if negativeType.IsValid() {
			t.Error("negative value should be invalid")
		}
	})

	t.Run("large positive value is invalid", func(t *testing.T) {
		largeType := AccountType(999)
		if largeType.IsValid() {
			t.Error("large positive value should be invalid")
		}
	})
}

// =============================================================================
// AccountType Boolean Methods Tests
// =============================================================================

func TestAccountType_BooleanMethods(t *testing.T) {
	t.Run("Parent.IsParent returns true", func(t *testing.T) {
		if !AccountTypeParent.IsParent() {
			t.Error("Parent.IsParent() should return true")
		}
	})

	t.Run("Parent.IsChild returns false", func(t *testing.T) {
		if AccountTypeParent.IsChild() {
			t.Error("Parent.IsChild() should return false")
		}
	})

	t.Run("Child.IsParent returns false", func(t *testing.T) {
		if AccountTypeChild.IsParent() {
			t.Error("Child.IsParent() should return false")
		}
	})

	t.Run("Child.IsChild returns true", func(t *testing.T) {
		if !AccountTypeChild.IsChild() {
			t.Error("Child.IsChild() should return true")
		}
	})

	t.Run("Unspecified.IsParent returns false", func(t *testing.T) {
		if AccountTypeUnspecified.IsParent() {
			t.Error("Unspecified.IsParent() should return false")
		}
	})

	t.Run("Unspecified.IsChild returns false", func(t *testing.T) {
		if AccountTypeUnspecified.IsChild() {
			t.Error("Unspecified.IsChild() should return false")
		}
	})

	t.Run("unknown type.IsParent returns false", func(t *testing.T) {
		unknownType := AccountType(999)
		if unknownType.IsParent() {
			t.Error("unknown type.IsParent() should return false")
		}
	})

	t.Run("unknown type.IsChild returns false", func(t *testing.T) {
		unknownType := AccountType(999)
		if unknownType.IsChild() {
			t.Error("unknown type.IsChild() should return false")
		}
	})
}

// =============================================================================
// AccountType.RequiresDID Tests
// =============================================================================

func TestAccountType_RequiresDID(t *testing.T) {
	t.Run("Parent requires DID", func(t *testing.T) {
		if !AccountTypeParent.RequiresDID() {
			t.Error("Parent should require DID")
		}
	})

	t.Run("Child does not require DID", func(t *testing.T) {
		if AccountTypeChild.RequiresDID() {
			t.Error("Child should not require DID")
		}
	})

	t.Run("Unspecified does not require DID", func(t *testing.T) {
		if AccountTypeUnspecified.RequiresDID() {
			t.Error("Unspecified should not require DID")
		}
	})

	t.Run("unknown type does not require DID", func(t *testing.T) {
		unknownType := AccountType(999)
		if unknownType.RequiresDID() {
			t.Error("unknown type should not require DID")
		}
	})
}

// =============================================================================
// AccountType.RequiresParent Tests
// =============================================================================

func TestAccountType_RequiresParent(t *testing.T) {
	t.Run("Child requires parent", func(t *testing.T) {
		if !AccountTypeChild.RequiresParent() {
			t.Error("Child should require parent")
		}
	})

	t.Run("Parent does not require parent", func(t *testing.T) {
		if AccountTypeParent.RequiresParent() {
			t.Error("Parent should not require parent")
		}
	})

	t.Run("Unspecified does not require parent", func(t *testing.T) {
		if AccountTypeUnspecified.RequiresParent() {
			t.Error("Unspecified should not require parent")
		}
	})

	t.Run("unknown type does not require parent", func(t *testing.T) {
		unknownType := AccountType(999)
		if unknownType.RequiresParent() {
			t.Error("unknown type should not require parent")
		}
	})
}

// =============================================================================
// AccountType.Validate Tests
// =============================================================================

func TestAccountType_Validate(t *testing.T) {
	t.Run("valid Parent type passes", func(t *testing.T) {
		err := AccountTypeParent.Validate()
		if err != nil {
			t.Errorf("Parent.Validate() should pass, got error: %v", err)
		}
	})

	t.Run("valid Child type passes", func(t *testing.T) {
		err := AccountTypeChild.Validate()
		if err != nil {
			t.Errorf("Child.Validate() should pass, got error: %v", err)
		}
	})

	t.Run("Unspecified type fails", func(t *testing.T) {
		err := AccountTypeUnspecified.Validate()
		if err == nil {
			t.Error("Unspecified.Validate() should fail")
		}
	})

	t.Run("unknown type fails", func(t *testing.T) {
		unknownType := AccountType(999)
		err := unknownType.Validate()
		if err == nil {
			t.Error("unknown type.Validate() should fail")
		}
	})

	t.Run("zero value fails", func(t *testing.T) {
		var zeroType AccountType
		err := zeroType.Validate()
		if err == nil {
			t.Error("zero value.Validate() should fail")
		}
	})

	t.Run("error has correct kind", func(t *testing.T) {
		err := AccountTypeUnspecified.Validate()
		if err == nil {
			t.Fatal("expected error")
		}
		var ae *AccountError
		if !errors.As(err, &ae) {
			t.Error("error should be AccountError")
		}
		if ae.Kind != ErrKindInvalidAccount {
			t.Errorf("error Kind = %v, want %v", ae.Kind, ErrKindInvalidAccount)
		}
	})
}

// =============================================================================
// AccountType Consistency Tests
// =============================================================================

func TestAccountType_Consistency(t *testing.T) {
	t.Run("IsParent implies RequiresDID", func(t *testing.T) {
		for _, at := range []AccountType{AccountTypeUnspecified, AccountTypeParent, AccountTypeChild} {
			if at.IsParent() && !at.RequiresDID() {
				t.Errorf("if IsParent() then RequiresDID() should be true for %v", at)
			}
		}
	})

	t.Run("IsChild implies RequiresParent", func(t *testing.T) {
		for _, at := range []AccountType{AccountTypeUnspecified, AccountTypeParent, AccountTypeChild} {
			if at.IsChild() && !at.RequiresParent() {
				t.Errorf("if IsChild() then RequiresParent() should be true for %v", at)
			}
		}
	})

	t.Run("IsParent and IsChild are mutually exclusive", func(t *testing.T) {
		for _, at := range []AccountType{AccountTypeUnspecified, AccountTypeParent, AccountTypeChild, AccountType(999)} {
			if at.IsParent() && at.IsChild() {
				t.Errorf("IsParent() and IsChild() should be mutually exclusive for %v", at)
			}
		}
	})
}
