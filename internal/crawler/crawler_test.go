package crawler

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
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

func TestNewCrawler(t *testing.T) {
	c, err := NewCrawler()
	if c == nil || err != nil {
		t.Errorf("initializing crawler failed: %v", err)
	}
	c, err = NewCrawler(&http.Client{})
	if c == nil || err != nil {
		t.Errorf("initializing crawler failed: %v", err)
	}
	c, err = NewCrawler(&http.Client{}, &http.Client{})
	if c != nil || err != ErrMultipleClient {
		t.Errorf("failed to prevent multiple http clients")
	}
}

func TestRetrieveProduct_PageNotFound(t *testing.T) {
	client := &MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewReader([]byte{})),
			}, nil
		},
	}
	c, err := NewCrawler(client)
	if c == nil || err != nil {
		t.Errorf("initializing crawler failed: %v", err)
	}

	r := c.RetrieveProduct("1000000")
	if r.Err == nil {
		t.Errorf("Err != nil when page is not available")
	}
}

func TestRetrieveProduct_ProductNotFound(t *testing.T) {
	client := getMockClientwithFile("./test/failed.html")
	c, err := NewCrawler(client)
	if c == nil || err != nil {
		t.Errorf("initializing crawler failed: %v", err)
	}

	r := c.RetrieveProduct("1000000")
	if r.Err == nil {
		t.Errorf("Should not be no error")
	}
	if r.Err != PRODUCT_NOT_FOUND {
		t.Errorf("Found product that should be not available")
	}
}

func TestRetrieveProduct_Success(t *testing.T) {
	client := getMockClientwithFile("./test/success.html")
	c, err := NewCrawler(client)
	if c == nil || err != nil {
		t.Errorf("initializing crawler failed: %v", err)
	}

	r := c.RetrieveProduct(mockResult.ProductCode)
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
