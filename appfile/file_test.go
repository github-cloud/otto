package appfile

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mitchellh/copystructure"
)

func TestFileActiveInfrastructure(t *testing.T) {
	cases := []struct {
		File   string
		Result string
	}{
		{
			"file-active-infra-basic.hcl",
			"aws",
		},
	}

	for _, tc := range cases {
		path := filepath.Join("./test-fixtures", tc.File)
		actual, err := ParseFile(path)
		if err != nil {
			t.Fatalf("file: %s\n\n%s", tc.File, err)
			continue
		}

		infra := actual.ActiveInfrastructure()
		if infra.Name != tc.Result {
			t.Fatalf("file: %s\n\n%s", tc.File, infra.Name)
		}
	}
}

func TestFileMerge(t *testing.T) {
	cases := map[string]struct {
		One, Two, Three *File
	}{
		"ID": {
			One: &File{
				ID: "foo",
			},
			Two: &File{
				ID: "bar",
			},
			Three: &File{
				ID: "bar",
			},
		},

		"Path": {
			One: &File{
				Path: "foo",
			},
			Two: &File{
				Path: "bar",
			},
			Three: &File{
				Path: "bar",
			},
		},

		"Application": {
			One: &File{
				Application: &Application{
					Name: "foo",
				},
			},
			Two: &File{
				Application: &Application{
					Type: "foo",
				},
			},
			Three: &File{
				Application: &Application{
					Name: "foo",
					Type: "foo",
				},
			},
		},

		"Application (no merge)": {
			One: &File{
				Application: &Application{
					Name: "foo",
				},
			},
			Two: &File{},
			Three: &File{
				Application: &Application{
					Name: "foo",
				},
			},
		},

		"Application (detect)": {
			One: &File{
				Application: &Application{
					Name:   "foo",
					Detect: true,
				},
			},
			Two: &File{
				Application: &Application{
					Type:   "foo",
					Detect: false,
				},
			},
			Three: &File{
				Application: &Application{
					Name:   "foo",
					Type:   "foo",
					Detect: false,
				},
			},
		},

		"Application (version)": {
			One: &File{
				Application: &Application{
					Name: "foo",
				},
			},
			Two: &File{
				Application: &Application{
					VersionRaw: "1.2.3",
				},
			},
			Three: &File{
				Application: &Application{
					Name:       "foo",
					VersionRaw: "1.2.3",
				},
			},
		},

		"Infra (no merge)": {
			One: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
					},
				},
			},
			Two: &File{},
			Three: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
					},
				},
			},
		},

		"Infra (add)": {
			One: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
					},
				},
			},
			Two: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "google",
					},
				},
			},
			Three: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
					},
					&Infrastructure{
						Name: "google",
					},
				},
			},
		},

		"Infra (override)": {
			One: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
					},
				},
			},
			Two: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
					},
				},
			},
			Three: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
					},
				},
			},
		},

		"Foundations (none)": {
			One: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
						Foundations: []*Foundation{
							&Foundation{
								Name: "consul",
							},
						},
					},
				},
			},
			Two: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
					},
				},
			},
			Three: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
						Foundations: []*Foundation{
							&Foundation{
								Name: "consul",
							},
						},
					},
				},
			},
		},

		"Foundations (override)": {
			One: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
						Foundations: []*Foundation{
							&Foundation{
								Name: "consul",
							},
						},
					},
				},
			},
			Two: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
						Foundations: []*Foundation{
							&Foundation{
								Name: "tubes",
							},
						},
					},
				},
			},
			Three: &File{
				Infrastructure: []*Infrastructure{
					&Infrastructure{
						Name: "aws",
						Foundations: []*Foundation{
							&Foundation{
								Name: "tubes",
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		if err := tc.One.Merge(tc.Two); err != nil {
			t.Fatalf("%s: %s", name, err)
		}

		if !reflect.DeepEqual(tc.One, tc.Three) {
			t.Fatalf("%s:\n\n%#v\n\n%#v", name, tc.One, tc.Three)
		}
	}
}

func TestFileDeepCopy(t *testing.T) {
	f, err := ParseFile(filepath.Join("./test-fixtures", "basic.hcl"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	f2, err := copystructure.Copy(f)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(f, f2) {
		t.Fatalf("bad:\n\n%#v\n\n%#v", f, f2)
	}
}

func TestFileConfigHash(t *testing.T) {
	cases := []struct {
		One, Two string
		Match    bool
	}{
		{
			"basic.hcl",
			"basic.hcl",
			true,
		},

		{
			"basic.hcl",
			"basic-diff.hcl",
			false,
		},
	}

	for _, tc := range cases {
		path1 := filepath.Join("./test-fixtures/config-hash", tc.One)
		path2 := filepath.Join("./test-fixtures/config-hash", tc.Two)
		actual1, err := ParseFile(path1)
		if err != nil {
			t.Fatalf("file: %s\n\n%s", path1, err)
			continue
		}
		actual2, err := ParseFile(path2)
		if err != nil {
			t.Fatalf("file: %s\n\n%s", path2, err)
			continue
		}

		v1, v2 := actual1.ConfigHash(), actual2.ConfigHash()
		if v1 == 0 {
			t.Fatalf("file: %s, zero hash", tc.One)
		}
		if (v1 == v2) != tc.Match {
			t.Fatalf("file:\n%s\n%s\n\n%#v", tc.One, tc.Two, tc.Match)
		}
	}
}

func TestApplicationVersion(t *testing.T) {
	cases := []struct {
		File   string
		Result string
	}{
		{
			"app-version.hcl",
			"1.0.0",
		},
	}

	for _, tc := range cases {
		path := filepath.Join("./test-fixtures", tc.File)
		actual, err := ParseFile(path)
		if err != nil {
			t.Fatalf("file: %s\n\n%s", tc.File, err)
			continue
		}

		vsn := actual.Application.Version()
		if vsn.String() != tc.Result {
			t.Fatalf("file: %s\n\n%s", tc.File, vsn)
		}
	}
}
