package validate

import "strings"

// ValidateEmail: empty is false, exactly one @, local and domain not empty,
// no whitespace. I also reject trailing/leading dots in the domain,
// consecutive dots, and any non-ascii rune.
func ValidateEmail(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t\n\r") {
		return false
	}
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	if strings.Count(s, "@") != 1 {
		return false
	}
	local, domain, _ := strings.Cut(s, "@")
	if local == "" || domain == "" {
		return false
	}
	if strings.Contains(local, "..") || strings.Contains(domain, "..") {
		return false
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	return true
}

// ValidatePhone: "+" then digits, or digits with optional dashes.
// No letters, no spaces, no parentheses. 10 to 15 digits.
func ValidatePhone(s string) bool {
	if s == "" {
		return false
	}
	rest := s
	if strings.HasPrefix(s, "+") {
		rest = s[1:]
	}
	if rest == "" {
		return false
	}
	digits := 0
	for _, r := range rest {
		if r == '-' {
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
		digits++
	}
	return digits >= 10 && digits <= 15
}
