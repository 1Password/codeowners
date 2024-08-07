package codeowners

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// defaultSection returns a Section with default values.
func defaultSection(s Section) Section {
	if s.Approvals == 0 {
		s.Approvals = 1
	}

	return s
}

func Test_parseSection(t *testing.T) {
	tests := []*struct {
		name    string
		section string
		want    Section
		wantErr bool
	}{
		{
			name:    "should parse an optional section",
			section: "^[Optional Section]",
			want: Section{
				Name:     "Optional Section",
				Optional: true,
			},
		},
		{
			name:    "should parse a section with approvals",
			section: "[Section Name][2]",
			want:    Section{Name: "Section Name", Approvals: 2},
		},
		{
			name:    "should parse a section with approvals and default owner",
			section: "[Section Name][2] @default.owner",
			want: Section{
				Name:      "Section Name",
				Approvals: 2,
				Owners:    []Owner{{Value: "default.owner", Type: UsernameOwner}},
			},
		},
		{
			name:    "should parse a section with approvals and multiple default owners",
			section: "[Section Name][2] @default.owner @people/gitlab-team",
			want: Section{
				Name:      "Section Name",
				Approvals: 2,
				Owners: []Owner{
					{Value: "default.owner", Type: UsernameOwner},
					{Value: "people/gitlab-team", Type: TeamOwner},
				},
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseSection(tt.section, parseOptions{ownerMatchers: DefaultOwnerMatchers})
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, defaultSection(tt.want), *got)
		})
	}
}

func Test_addFromSection(t *testing.T) {
	type args struct {
		r *Rule
		s *Section
	}
	tests := []*struct {
		name string
		args args
		want Rule
	}{
		{
			name: "should add default owners to a rule",
			args: args{
				r: &Rule{},
				s: &Section{
					Owners: []Owner{{Value: "default.owner", Type: UsernameOwner}},
				},
			},
			want: Rule{
				Owners: []Owner{{Value: "default.owner", Type: UsernameOwner}},
			},
		},
		{
			name: "should not add default owners if a rule already has owners",
			args: args{
				r: &Rule{
					Owners: []Owner{{Value: "user", Type: UsernameOwner}},
				},
				s: &Section{
					Owners: []Owner{{Value: "default.owner", Type: UsernameOwner}},
				},
			},
			want: Rule{
				Owners: []Owner{{Value: "user", Type: UsernameOwner}},
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			addFromSection(tt.args.r, tt.args.s)
			assert.Equal(t, *tt.args.r, tt.want)
		})
	}
}
