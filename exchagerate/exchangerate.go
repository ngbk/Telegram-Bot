package exchagerate

import (
	"fmt"
	"github.com/gocolly/colly"
)

type Fiat struct {
	name  string
	price string
}

type CryptoCurrency struct {
	name  string
	price string
}

func ParseFiat() map[string]string {
	res := make(map[string]string, 100)
	url := "https://nationalbank.kz/ru/exchangerates/ezhednevnye-oficialnye-rynochnye-kursy-valyut"

	coll := colly.NewCollector()
	coll.OnHTML("tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			res[el.ChildText("td:nth-child(3)")[0:3]] = el.ChildText(`td:nth-child(4)`)
		})
	})

	coll.OnError(func(response *colly.Response, err error) {
		fmt.Println(err.Error())
	})
	_ = coll.Visit(url)
	return res
}

func ParseCrypto() map[string]string {
	res := make(map[string]string, 100)

	url := "https://coinmarketcap.com/all/views/all/"

	coll := colly.NewCollector()

	coll.OnHTML(".cmc-table__table-wrapper-outer .cmc-table-row", func(e *colly.HTMLElement) {
		res[e.ChildText(".cmc-table__column-name--name")] = e.ChildText(".cmc-link span")
	})
	_ = coll.Visit(url)

	return res
}
