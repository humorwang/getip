package main

import (
	"flag"
	"github.com/humorwang/getip/src/realip"
	"fmt"
	"log"
	"strconv"
	"net"
	"github.com/oschwald/geoip2-golang"
	"github.com/gin-gonic/gin"
)

var port string

func getIpInfo(ipStr string)(string) {
	isPrivate, err := realip.IsPrivateAddress(ipStr)
	if err != nil {
		return "局域网地址检查失败"
	}
	if isPrivate {
		return "局域网地址"
	}
	db, err := geoip2.Open("./geolite2/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
		return "获取ip地址信息失败"
	}
	defer db.Close()
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(ipStr)
	record, err := db.City(ip)
	if err != nil {
		log.Fatal(err)
		return "获取ip地址信息失败"
	}
	countryCode :=  record.Country.IsoCode
	cityName := record.City.Names["en"]
	countryName := record.Country.Names["en"]
	if countryCode == "CN" {
		cityName = record.City.Names["zh-CN"]
		if len(record.Subdivisions) > 0 {
			cityName = fmt.Sprintf("%v %v", record.Subdivisions[0].Names["zh-CN"], cityName)
		}
		countryName = record.Country.Names["zh-CN"]
	} else {
		if len(record.Subdivisions) > 0 {
			cityName = fmt.Sprintf("%v %v", record.Subdivisions[0].Names["en"], cityName)
		}
	}
	// fmt.Printf("Portuguese (BR) city name: %v\n", record.City.Names["zh-CN"])
	// if len(record.Subdivisions) > 0 {
	// 	fmt.Printf("English subdivision name: %v\n", record.Subdivisions[0].Names["en"])
	// }
	// fmt.Printf("Russian country name: %v\n", record.Country.Names["en"])
	// fmt.Printf("ISO country code: %v\n", record.Country.IsoCode)
	// fmt.Printf("Time zone: %v\n", record.Location.TimeZone)
	// fmt.Printf("Coordinates: %v, %v\n", record.Location.Latitude, record.Location.Longitude)
	return fmt.Sprintf("%v %v %v", cityName, countryName, countryCode)
}

func response(c *gin.Context) {
	c.Request.ParseForm()
	c.Request.ParseMultipartForm(33554432)
	response_code := 200
	format := c.DefaultQuery("format", "text")
	http_code := c.DefaultQuery("http_code", "200")
	if value, err := strconv.Atoi(http_code); err == nil {
		response_code = value
	}
	ip := c.ClientIP()
	RealIp := realip.FromRequest(c.Request)
	ipInfo := getIpInfo(RealIp)
	response_json := make(map[string]interface{})
	response_json["ClientIp"] = ip
	response_json["RealIp"] = RealIp
	response_json["IpAddress"] = ipInfo

	// log.Printf("\n============================================================================\n"+
	// 	"IP:%s\n"+
	// 	"X-Forwarded-For:%s\n"+
	// 	"X-Real-Ip:%s\n"+
	// 	"X-Forwarded-Host:%s\n"+
	// 	"RemoteAddr:%s\n"+
	// 	"Content-Type:%s\n"+
	// 	"IpInfo:%s\n",
	// 	c.ClientIP(),
	// 	c.Request.Header.Get("X-Forwarded-For"),
	// 	c.Request.Header.Get("X-Real-Ip"),
	// 	c.Request.Header.Get("X-Forwarded-Host:"),
	// 	c.Request.RemoteAddr,
	// 	c.Request.Header.Get("Content-Type"),
	// 	getIpInfo(ip))
	log.Printf("RealIp:%s IpInfo:%s", RealIp, getIpInfo(RealIp))
	if format == "text" {
		c.String(response_code, "%s %s", RealIp, ipInfo)
	} else {
		c.JSON(response_code, response_json)
	}
}

func init() {
	flag.StringVar(&port, "port", ":8080", "端口")
}
func main() {
	flag.Parse()
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	v1 := r.Group("/")
	v1.GET("/*router", response)
	v1.HEAD("/*router", response)
	v1.POST("/*router", response)
	v1.PUT("/*router", response)
	v1.DELETE("/*router", response)
	v1.OPTIONS("/*router", response)
	r.Run(port)
}
