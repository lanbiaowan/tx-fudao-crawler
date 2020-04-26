module crawler

go 1.13

require (
	github.com/bitly/go-simplejson v0.5.0
	github.com/gin-gonic/gin v1.6.2
	github.com/lilien1010/tx-fudao-crawler v0.0.0-20200425071749-4f13b6364415
	github.com/lilien1010/tx-fudao-crawler/crawler v0.0.0-00010101000000-000000000000 // indirect
)

replace github.com/lilien1010/tx-fudao-crawler => ../../tx-fudao-crawler

replace github.com/lilien1010/tx-fudao-crawler/crawler => ../../tx-fudao-crawler/crawler
