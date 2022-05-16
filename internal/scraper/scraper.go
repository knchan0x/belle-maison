package scraper

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36"
	baseURL    = "https://www.bellemaison.jp/shop/commodity/0000/"
)

type Result struct {
	ProductCode string
	Product     *Product
	Err         error
}

type Product struct {
	Name   string
	Styles []Style
}

type Style struct {
	StyleCode string
	ImageUrl  string
	Colour    string
	Size      string
	Price     uint
	Stock     uint
}

// Scraper
type Scraper interface {
	Scraping(productCodes ...string) []*Result
}

// scraper implements Scraper interface
type scraper struct {
	httpClient HTTPClient
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var ErrMultipleClient = errors.New("more than one http client assigned")

// NewScraper return a Scraper instance,
// default http client will be used if no httpClient provided
func NewScraper(httpClient ...HTTPClient) (Scraper, error) {
	if len(httpClient) > 1 {
		return nil, ErrMultipleClient
	}

	var client HTTPClient
	if len(httpClient) == 0 {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: time.Second * 60,
		}
	} else {
		client = httpClient[0]
	}

	return &scraper{httpClient: client}, nil
}

// http response
type response struct {
	id   string
	data []byte
	err  error
}

// ScrapingProducts fetches and parses multiple products from the site
func (c *scraper) Scraping(productCodes ...string) []*Result {
	size := len(productCodes)

	if size <= 0 {
		return nil
	}

	respCh := make(chan *response, size)
	resultCh := make(chan *Result, size)
	go func(respCh <-chan *response, resultCh chan<- *Result) {
		var wg sync.WaitGroup
		for resp := range respCh {
			wg.Add(1)
			go func(resp response, resultCh chan<- *Result) {
				defer wg.Done()
				result := Result{
					ProductCode: resp.id,
				}
				if resp.err == nil {
					result.Product, result.Err = parseHTML(resp.data)

				} else {
					result.Err = resp.err
				}

				resultCh <- &result
			}(*resp, resultCh)
		}
		wg.Wait()
		close(resultCh)
	}(respCh, resultCh)

	for _, id := range productCodes {
		resp := response{
			id: id,
		}
		data, err := fetch(c.httpClient, baseURL+id)
		if err != nil {
			resp.err = err
		} else {
			resp.data = data
		}
		respCh <- &resp
	}
	close(respCh)

	results := make([]*Result, size)
	idx := 0
	for result := range resultCh {
		results[idx] = result
		idx++
	}

	return results
}

// fetch fetches web page from the site
func fetch(httpClient HTTPClient, link string) ([]byte, error) {
	// Set header
	req, _ := http.NewRequest(http.MethodGet, link, nil)
	req.Header.Set("user-agent", USER_AGENT)

	// HTTP Request
	res, err := httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return []byte{}, fmt.Errorf("%d: %s", res.StatusCode, res.Status)
	}

	val, _ := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return val, nil
}

var (
	PRODUCT_NOT_FOUND = errors.New("Product Not Found")
)

// parseHTML converts html page to Product struct
func parseHTML(html []byte) (*Product, error) {

	page, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}

	title := ""
	page.Find("h1[class='title']").Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})

	// product has been removed
	if title == "お探しの商品が見つかりません" {
		return nil, PRODUCT_NOT_FOUND
	}

	newProduct := &Product{
		Styles: []Style{},
	}

	// get product name
	page.Find("h1[class='product-name text-weight-bold']").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			newProduct.Name = s.Text()
		}
	})

	// find the target items
	var colour, size, current, stock, sku string
	page.Find("#commodityStandardAreaMessage").Parent().Each(func(i int, s *goquery.Selection) {
		s.Find(".standard-info").Each(func(i int, s *goquery.Selection) {
			colour, _ = s.Attr("data-standard-detail2")

			// check colour info
			if colour == "-" {
				colour = "Standard"
			}

			size, _ = s.Attr("data-standard-detail1")

			// check size info
			if size == "-" {
				size = "Standard"
			}

			// in case it has additional size info
			if detail, exist := s.Attr("data-standard-detail12"); exist {
				if detail != "-" {
					size = size + "/" + detail
				}
			}

			current, _ = s.Attr("data-price")
			stock, _ = s.Attr("data-stock-status")
			sku, _ = s.Attr("data-nucleus-sku-code")

			currentPrice, err := strconv.ParseUint(strings.ReplaceAll(current, ",", ""), 10, 64)
			if err != nil {
				currentPrice = 0
			}
			stockNo := parseStock(stock)

			image := ""
			s.Siblings().Find(".variation-list_item").Each(func(i int, s *goquery.Selection) {
				s.Find("input[name='color']").Each(func(i int, s *goquery.Selection) {
					itemColor, _ := s.Attr("data-name")
					if colour == itemColor {
						image, _ = s.Attr("data-img")
					}
				})
			})

			if image == "" {
				s.Siblings().Find(".variation-check-radio").Each(func(i int, s *goquery.Selection) {
					s.Find("input[name='color']").Each(func(i int, s *goquery.Selection) {
						itemColor, _ := s.Attr("data-name")
						if colour == itemColor {
							image, _ = s.Attr("data-img")
						}
					})
				})
			}

			if image == "" {
				image = fmt.Sprintf("https://pic2.bellemaison.jp/shop/cms/images/0000/catalog/%s/%s_h1_001.jpg", sku[:7], sku[:7])
			}

			newStyle := Style{
				StyleCode: sku[7:],
				ImageUrl:  image,
				Colour:    colour,
				Size:      size,
				Price:     uint(currentPrice),
				Stock:     stockNo,
			}
			newProduct.Styles = append(newProduct.Styles, newStyle)
		})
	})

	return newProduct, nil
}

func parseStock(description string) (stock uint) {
	switch description {
	case "在庫あり":
		stock = 99
	case "売り切れ":
		fallthrough
	case "販売停止":
		fallthrough
	case "売り切れ（再入荷なし）":
		stock = 0
	default:
		if strings.Contains(description, "在庫：") {
			temp := strings.Split(description, "：")
			tempNo, err := strconv.ParseUint(temp[1], 10, 64)
			if err != nil {
				stock = 0
			} else {
				stock = uint(tempNo)
			}
		}

		if strings.Contains(description, "入荷予定") {
			stock = 0
		}
	}
	return stock
}
