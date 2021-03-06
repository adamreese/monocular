package helpers

import (
	"fmt"
	"strings"

	"github.com/helm/monocular/src/api/pkg/swagger/models"
	"gopkg.in/yaml.v2"
)

// IsYAML checks for valid YAML
func IsYAML(b []byte) bool {
	var yml map[string]interface{}
	ret := yaml.Unmarshal(b, &yml)
	return ret == nil
}

// ParseYAMLRepo converts a YAML representation of a repo
// to a slice of versioned charts
func ParseYAMLRepo(rawYAML []byte) ([]models.ChartVersion, error) {
	repo := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(rawYAML, &repo); err != nil {
		return nil, err
	}
	var charts []models.ChartVersion
	for chartVersion := range repo {
		cV := repo[chartVersion]
		// we drop the error response because we can never enter into this error state:
		// 1. "repo" is statically produced above
		// 2. the key value at "cV" is a resilient, yaml-marshallable interface: it was unmarshalled from yaml above
		c, _ := yaml.Marshal(&cV)
		var chart models.ChartVersion
		if err := yaml.Unmarshal(c, &chart); err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}
	return charts, nil
}

// MakeChartResource composes a Resource type that represents a repo+chart
func MakeChartResource(chart models.ChartVersion, repo, version string) models.Resource {
	var ret models.Resource
	ret.Type = StrToPtr("chart")
	ret.ID = StrToPtr(fmt.Sprintf("%s/%s", repo, *chart.Name))
	ret.Links = &models.ChartResourceLinks{
		Latest: StrToPtr(fmt.Sprintf("/v1/charts/%s/%s/%s", repo, *chart.Name, version)),
	}
	ret.Attributes = &models.ChartResourceAttributes{
		Repo:        &repo,
		Name:        chart.Name,
		Description: chart.Description,
		Created:     chart.Created,
		Home:        chart.Home,
	}
	return ret
}

// GetLatestChartVersion returns the most recent version from a slice of versioned charts
func GetLatestChartVersion(charts []models.ChartVersion, name string) (models.ChartVersion, error) {
	latest := "0.0.0"
	var ret models.ChartVersion
	for _, chart := range charts {
		if *chart.Name == name {
			newest, err := newestSemVer(latest, *chart.Version)
			if err != nil {
				return models.ChartVersion{}, err
			}
			latest = newest
			if latest == *chart.Version {
				ret = chart
			}
		}
	}
	return ret, nil
}

// newestSemVer returns the newest (largest) semver string
func newestSemVer(v1 string, v2 string) (string, error) {
	v1Slice := strings.Split(v1, ".")
	if len(v1Slice) != 3 {
		return "", semverStringError(v1)
	}
	v2Slice := strings.Split(v2, ".")
	if len(v2Slice) != 3 {
		return "", semverStringError(v2)
	}
	for i, subVer1 := range v1Slice {
		if v2Slice[i] > subVer1 {
			return v2, nil
		} else if subVer1 > v2Slice[i] {
			return v1, nil
		}
	}
	return v1, nil
}

// semverStringError returns a bad semver string error
func semverStringError(v string) error {
	return fmt.Errorf("%s is not a semver-compatible string", v)
}

// Int64ToPtr converts an int64 to an *int64
func Int64ToPtr(n int64) *int64 {
	return &n
}

// StrToPtr converts a string to a *string
func StrToPtr(s string) *string {
	return &s
}
