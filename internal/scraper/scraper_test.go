package scraper

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{}, nil
}

// TODO: combine with TestNewScraper
// func TestNewCrawler(t *testing.T) {
// 	c, err := NewScraper()
// 	if c == nil || err != nil {
// 		t.Errorf("initializing crawler failed: %v", err)
// 	}
// 	c, err = NewScraper(&http.Client{})
// 	if c == nil || err != nil {
// 		t.Errorf("initializing crawler failed: %v", err)
// 	}
// 	c, err = NewScraper(&http.Client{}, &http.Client{})
// 	if c != nil || err != ErrMultipleClient {
// 		t.Errorf("failed to prevent multiple http clients")
// 	}
// }

func TestRetrieveProduct_PageNotFound(t *testing.T) {
	client := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		},
	}
	c, err := NewScraper(client)
	if c == nil || err != nil {
		t.Errorf("initializing scraper failed: %v", err)
	}

	r := c.Scraping("1000000")[0]
	if r.Err == nil {
		t.Errorf("Err != nil when page is not available")
	}
}

func TestRetrieveProduct_ProductNotFound(t *testing.T) {
	client := getMockClientwithFile("./test/failed.html")
	c, err := NewScraper(client)
	if c == nil || err != nil {
		t.Errorf("initializing scraper failed: %v", err)
	}

	r := c.Scraping("1000000")[0]
	if r.Err == nil {
		t.Errorf("Should not be no error")
	}
	if r.Err != PRODUCT_NOT_FOUND {
		t.Errorf("Found product that should be not available")
	}
}

func TestRetrieveProduct_Success(t *testing.T) {
	client := getMockClientwithFile("./test/success.html")
	c, err := NewScraper(client)
	if c == nil || err != nil {
		t.Errorf("initializing scraper failed: %v", err)
	}

	r := c.Scraping(mockResult.ProductCode)[0]
	if r.ProductCode != mockResult.ProductCode {
		t.Errorf("incorrect product code")
	}
	if r.Err != nil {
		t.Errorf("Error: %v", err)
	}
	if r.Product.Name != mockResult.Product.Name {
		t.Errorf("Error: %v", err)
	}
	for i, mockStyle := range mockResult.Product.Styles {
		if r.Product.Styles[i].StyleCode != mockStyle.StyleCode {
			t.Errorf("style not match")
		}
		if r.Product.Styles[i].ImageUrl != mockStyle.ImageUrl {
			t.Errorf("imageUrl not match")
		}
		if r.Product.Styles[i].Colour != mockStyle.Colour {
			t.Errorf("colour not match")
		}
		if r.Product.Styles[i].Size != mockStyle.Size {
			t.Errorf("size not match")
		}
		if r.Product.Styles[i].Price != mockStyle.Price {
			t.Errorf("price not match")
		}
		if r.Product.Styles[i].Stock != mockStyle.Stock {
			t.Errorf("stock not match")
		}
	}
}

func getMockClientwithFile(name string) *MockClient {
	getMockHtmlPage()

	file, err := os.ReadFile(name)
	if err != nil {
		log.Panic("test file not available")
	}

	client := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(file)),
			}, nil
		},
	}
	return client
}

// need to match with the html file downloaded (./test/success.html)
var mockResult = &Result{
	ProductCode: "1129250",
	Product: &Product{
		Name: "シートマッサージャー",
		Styles: []Style{
			{StyleCode: "01001", ImageUrl: "https://pic2.bellemaison.jp/shop/cms/images/0000/catalog/1129250/1129250_h1_001.jpg", Colour: "Standard", Size: "Standard", Price: 6578, Stock: 4},
		},
	},
	Err: nil,
}

func getMockHtmlPage() {
	if err := download("./test/success.html", "https://www.bellemaison.jp/shop/commodity/0000/1129250"); err != nil {
		panic(err)
	}

	if err := download("./test/failed.html", "https://www.bellemaison.jp/shop/commodity/0000/1000000"); err != nil {
		panic(err)
	}
}

func download(filepath string, url string) error {

	// get data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// create file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// write to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func TestNewScraper(t *testing.T) {
	type args struct {
		httpClient []HTTPClient
	}
	tests := []struct {
		name    string
		args    args
		want    Scraper
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScraper(tt.args.httpClient...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScraper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScraper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scraper_ScrapingProducts(t *testing.T) {
	type fields struct {
		httpClient HTTPClient
	}
	type args struct {
		productCodes []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*Result
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &scraper{
				httpClient: tt.fields.httpClient,
			}
			if got := c.Scraping(tt.args.productCodes...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scraper.ScrapingProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}
