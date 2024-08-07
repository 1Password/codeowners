package codeowners

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var sectionRegexp = regexp.MustCompile(`^(?<optional>\^)?\[(?<name>[^\]\[\n]+)\](?:\[(?<approvals>\d+)\])?(?: )?(?<owners>.*)$`)

// Section represents a Gitlab-flavored CODEOWNERS section of a file.
//
// Format:
//
//		^ = Optional, denotes a section as optional
//		[2] = Optional number of approvals
//	 @default-owner = Optional, default owner for the section
//	 -------
//		^[Section Name][2] @default-owner
//
// For more information, see their documentation:
// https://docs.gitlab.com/ee/user/project/codeowners/#organize-code-owners-by-putting-them-into-sections
//
// When a rule is under a section, it will have defaults applied to it.
type Section struct {
	// Optional denotes if a section has turned rules into optional
	// (not-blocking) or not. Defaults to false.
	//
	// This is indicated by a ^ prefix.
	Optional bool `json:"optional,omitempty"`

	// Name is the name of the section.
	Name string `json:"name"`

	// Approvals is the number of approvals required to satisfy the
	// section. Defaults to 1.
	Approvals int `json:"approvals,omitempty"`

	// Owners are the default owners for any rule in this section.
	Owners []Owner `json:"owners"`
}

// parseSection parses a section from a string.
func parseSection(s string, opts parseOptions) (*Section, error) {
	matches := sectionRegexp.FindStringSubmatch(s)
	if matches == nil {
		return nil, ErrNoMatch
	}
	if len(matches) != 5 {
		return nil, fmt.Errorf("regexp return an unexpected number of matches: %d", len(matches))
	}

	optionalIdx := sectionRegexp.SubexpIndex("optional")
	nameIdx := sectionRegexp.SubexpIndex("name")
	approvalsIdx := sectionRegexp.SubexpIndex("approvals")
	ownersIdx := sectionRegexp.SubexpIndex("owners")

	approvals := 1
	if approvalsStr := matches[approvalsIdx]; approvalsStr != "" {
		var err error
		approvals, err = strconv.Atoi(approvalsStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse approvals from section as int: %w", err)
		}
	}

	section := &Section{
		Optional:  matches[optionalIdx] == "^",
		Name:      matches[nameIdx],
		Approvals: approvals,
	}

	// create a stub rule so we can parse the owners of the section, if
	// present
	ownersStr := matches[ownersIdx]
	if strings.TrimSpace(ownersStr) != "" {
		fakeRule := "/* " + ownersStr
		r, err := parseRule(fakeRule, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse section owners: %w", err)
		}

		section.Owners = r.Owners
	}

	return section, nil
}

// addFromSection adds the fields from a section to a rule based on
// Gitlab's documentation.
func addFromSection(r *Rule, s *Section) {
	r.Optional = s.Optional

	// Only add rules from the section if there isn't already some from
	// the rule itself, as per Gitlab's docs.
	//
	// https://docs.gitlab.com/ee/user/project/codeowners/#use-default-owners-and-optional-sections-together
	if len(r.Owners) == 0 && len(s.Owners) > 0 {
		r.Owners = s.Owners
	}
}
