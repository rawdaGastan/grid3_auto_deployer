package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var (
	DevNetwork  = "dev"
	QaNetwork   = "qa"
	TestNetwork = "test"
	MainNetwork = "main"

	// GraphQlURLs for graphql urls
	GraphQlURLs = map[string][]string{
		DevNetwork: {
			"https://graphql.dev.grid.tf/graphql",
			"https://graphql.02.dev.grid.tf/graphql",
		},
		TestNetwork: {
			"https://graphql.test.grid.tf/graphql",
			"https://graphql.02.test.grid.tf/graphql",
		},
		QaNetwork: {
			"https://graphql.qa.grid.tf/graphql",
			"https://graphql.02.qa.grid.tf/graphql",
		},
		MainNetwork: {
			"https://graphql.grid.tf/graphql",
			"https://graphql.02.grid.tf/graphql",
		},
	}
)

// GraphQl for tf graphql
type GraphQl struct {
	urls           []string
	activeStackIdx int
}

// NewGraphQl new tf graphql
func NewGraphQl(network string) (GraphQl, error) {
	if len(network) == 0 {
		return GraphQl{}, errors.New("network is required")
	}

	return GraphQl{urls: GraphQlURLs[network], activeStackIdx: 0}, nil
}

// ListContractsByTwinID returns contracts for a twinID
func (g *GraphQl) ListRegions(countries []string) ([]string, error) {
	options := fmt.Sprintf("(orderBy: region_ASC, where: {name_in: %q})", countries)
	countriesCount, err := g.getItemTotalCount("countries", options)
	if err != nil {
		return nil, err
	}

	countriesData, err := g.query(`query getRegions($countriesCount: Int!){
						countries(limit: $countriesCount) {
              region
            }
          }`,
		map[string]interface{}{
			"countriesCount": countriesCount,
		})
	if err != nil {
		return nil, err
	}

	countriesJSONData, err := json.Marshal(countriesData)
	if err != nil {
		return nil, err
	}

	var listCountries struct {
		Countries []struct {
			Region string
		}
	}
	err = json.Unmarshal(countriesJSONData, &listCountries)
	if err != nil {
		return nil, err
	}

	var regions []string
	for _, c := range listCountries.Countries {
		if !slices.Contains(regions, c.Region) {
			regions = append(regions, c.Region)
		}
	}

	return regions, nil
}

// getItemTotalCount return count of items
func (g *GraphQl) getItemTotalCount(itemName string, options string) (float64, error) {
	countBody := fmt.Sprintf(`query { items: %vConnection%v { count: totalCount } }`, itemName, options)
	requestBody := map[string]interface{}{"query": countBody}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return 0, err
	}

	bodyReader := bytes.NewReader(jsonBody)

	countResponse, err := g.httpPost(bodyReader)
	if err != nil {
		return 0, err
	}

	queryData, err := parseHTTPResponse(countResponse)
	if err != nil {
		return 0, err
	}

	countMap := queryData["data"].(map[string]interface{})
	countItems := countMap["items"].(map[string]interface{})
	count := countItems["count"].(float64)

	return count, nil
}

// query queries graphql
func (g *GraphQl) query(body string, variables map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	requestBody := map[string]interface{}{"query": body, "variables": variables}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return result, err
	}

	bodyReader := bytes.NewReader(jsonBody)

	resp, err := g.httpPost(bodyReader)
	if err != nil {
		return result, err
	}

	queryData, err := parseHTTPResponse(resp)
	if err != nil {
		return result, err
	}

	result = queryData["data"].(map[string]interface{})
	return result, nil
}

func parseHTTPResponse(resp *http.Response) (map[string]interface{}, error) {
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{}, err
	}

	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.Unmarshal(resBody, &data)
	if err != nil {
		return map[string]interface{}{}, err
	}

	if resp.StatusCode >= 400 {
		return map[string]interface{}{}, errors.Errorf("request failed with status code: %d with error %v", resp.StatusCode, data)
	}

	return data, nil
}

func (g *GraphQl) httpPost(body io.Reader) (*http.Response, error) {
	cl := &http.Client{
		Timeout: 10 * time.Second,
	}

	var (
		endpoint string
		reqErr   error
		resp     *http.Response
	)

	backoffCfg := backoff.WithMaxRetries(
		backoff.NewConstantBackOff(1*time.Millisecond),
		2,
	)

	err := backoff.RetryNotify(func() error {
		endpoint = g.urls[g.activeStackIdx]
		log.Debug().Str("url", endpoint).Msg("checking")

		resp, reqErr = cl.Post(endpoint, "application/json", body)
		if reqErr != nil &&
			(errors.Is(reqErr, http.ErrAbortHandler) ||
				errors.Is(reqErr, http.ErrHandlerTimeout) ||
				errors.Is(reqErr, http.ErrServerClosed)) {
			g.activeStackIdx = (g.activeStackIdx + 1) % len(g.urls)
			return reqErr
		}

		return nil
	}, backoffCfg, func(err error, _ time.Duration) {
		log.Error().Err(err).Msg("failed to connect to endpoint, retrying")
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to connect to endpoint")
	}

	return resp, reqErr
}
