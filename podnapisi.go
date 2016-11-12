package podnapisi

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	libxml2 "github.com/lestrrat/go-libxml2"
	"github.com/lestrrat/go-libxml2/xpath"
)

var ALL_LANGUAGES = "ALL"

type Subtitle struct {
	Title    string
	Releases []string
	Season   string
	Episode  string
	Language string
	URL      string
}

type ShowSearchParams struct {
	Name     string
	Season   string
	Episode  string
	Download string
	Language string
	Limit    int
}

func Search(params ShowSearchParams) ([]Subtitle, error) {
	subtitleData, err := searchSubtitles(params)
	if err != nil {
		return nil, err
	}
	subtitles, err := parseSubtitles(subtitleData)
	if err != nil {
		return nil, err
	}
	return subtitles, nil
}

func searchSubtitles(searchParams ShowSearchParams) ([]byte, error) {
	params := make(map[string]string)
	params["sK"] = searchParams.Name
	params["sTS"] = searchParams.Season
	params["sTE"] = searchParams.Episode

	if searchParams.Language != ALL_LANGUAGES {
		params["sL"] = searchParams.Language
	}
	params["sXML"] = "1"
	requestUrl := "https://www.podnapisi.net/subtitles/search/old?"
	queryString := ""

	requestParams := make([]string, 0)

	for key, value := range params {
		requestParams = append(requestParams, key+"="+url.QueryEscape(value))
	}

	queryString = strings.Join(requestParams, "&")
	fullUrl := requestUrl + queryString

	response, err := http.Get(fullUrl)

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func parseSubtitles(input []byte) ([]Subtitle, error) {
	d, err := libxml2.ParseString(string(input))
	if err != nil {
		return nil, err
	}
	ctx, err := xpath.NewContext(d)
	if err != nil {
		return nil, err
	}

	subtitles := make([]Subtitle, 0)

	subtitleNodes := xpath.NodeList(ctx.Find("//subtitle"))

	for _, subtitle := range subtitleNodes {
		subCtx, err := xpath.NewContext(subtitle)

		if err != nil {
			return nil, err
		}

		titleNode, err := subCtx.Find("./title")
		if err != nil {
			return nil, err
		}

		title := titleNode.String()
		urlNode := xpath.NodeList(subCtx.Find("./url"))
		languageNode, err := subCtx.Find("./language")
		if err != nil {
			return nil, nil
		}

		seasonNode, err := subCtx.Find("./tvSeason")
		if err != nil {
			return nil, nil
		}

		episodeNode, err := subCtx.Find("./tvEpisode")
		if err != nil {
			return nil, nil
		}

		language := languageNode.String()
		url := urlNode.NodeValue() + "/download"
		releases := xpath.NodeList(subCtx.Find(".//releases/release"))
		releaseCollection := make([]string, 0)
		season := seasonNode.String()
		episode := episodeNode.String()

		for _, release := range releases {
			releaseCollection = append(releaseCollection, release.NodeValue())
		}

		subtitles = append(subtitles,
			Subtitle{Title: title, Releases: releaseCollection, Season: season,
				Episode: episode, Language: language, URL: url})
	}

	return subtitles, nil
}
