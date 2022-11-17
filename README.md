# Gocrawler

## Features

- Fast crawling over given inputs
- Stdin/Stdout support tooling integration
- Integrated request throttler
- Support for advanced scopes

### Known limitations.
This crawler won't properly work with javascript framework that are not
server-side rendered since it does not feature a renderer.

## Installation

### Go Module

To install `GoCrawler` as a `GO` module:
```shell
$ go install -v github.com/m1dugh/gocrawler@latest
```

### Docker 
The docker hub image is coming soon...

## Usage

```shell
$ gocrawler -h
```

It displays the help menu.

```shell
Usage of gocrawler:
  -requests int
        Max requests per second (default -1)
  -scope string
        the file containing regexes for the scope
  -threads uint
        Number of max concurrent threads (default 10)
  -urls string
        The file containing all urls
```

> `GoCrawler` has to be used with a json scope file provided by the `-scope` file

```json
{
    "include": [
        "\\.hackerone\\.com",
        "api\\.hackerone\\.com"
    ],

    "exclude": [
        "^http://"
    ]
}
```

> This scope provides two arrays containing regexes.

- `include`: It contains all the regexes that has to be matched for the url
to be added.

- `exclude`: It contains a set of regexes to remove certain urls validated in
`include`.

### Running GoCrawler
To run `GoCrawler`, urls can be provided through a file or directly to stdin.

*Note: If both used, only stdin will be used.*

```shell
$ echo "https://hackerone.com" | gocrawler -scope scope.json

https://www.hackerone.com
https://consent.trustarc.com/notice?domain=hackerone.com&c=teconsent&js=nj¬iceType=bb>m=1
https://www.hackerone.com/sites/default/files/HAC_ARM_Organic
https://www.hackerone.com/security-at-beyond
https://www.hackerone.com/attack-resistance-assessment
https://hackerone.com/users/sign_in
https://www.hackerone.com/product/attack-surface-management
https://hackerone.com/hacktivity
https://hackerone.com/leaderboard/all-time
https://hackerone.com/directory/programs?order_direction=DESC&order_field=resolved_report_count
https://www.hackerone.com/resources
https://www.hackerone.com/customer-stories
https://www.hackerone.com/hackerone-community-blog
https://www.hackerone.com/solutions/cloud-security-solution
https://www.hackerone.com/solutions/attack-resistance-management
https://www.hackerone.com/product/bug-bounty-platform
https://www.hackerone.com/product/response-vulnerability-disclosure-program
https://www.hackerone.com/product/pentest
https://www.hackerone.com/product/overview
https://hackerone.com/leaderboard
https://www.hackerone.com/resources/customer-story/how-hired-builds-customer-trust-with-hackerone-pentest
https://www.hackerone.com/customer-hub/Hyatt
https://www.hackerone.com/node/10766
https://hackerone.com/assets/static/css/vendor.8b11831d.css
https://hackerone.com/assets/static/css/main.b2076af6.css
```
or
```shell
$ gocrawler -scope scope.json -urls urls.txt

https://www.hackerone.com
https://consent.trustarc.com/notice?domain=hackerone.com&c=teconsent&js=nj¬iceType=bb>m=1
https://www.hackerone.com/sites/default/files/HAC_ARM_Organic
https://www.hackerone.com/security-at-beyond
https://www.hackerone.com/attack-resistance-assessment
https://hackerone.com/users/sign_in
https://www.hackerone.com/product/attack-surface-management
https://hackerone.com/hacktivity
https://hackerone.com/leaderboard/all-time
https://hackerone.com/directory/programs?order_direction=DESC&order_field=resolved_report_count
https://www.hackerone.com/resources
https://www.hackerone.com/customer-stories
https://www.hackerone.com/hackerone-community-blog
https://www.hackerone.com/solutions/cloud-security-solution
https://www.hackerone.com/solutions/attack-resistance-management
https://www.hackerone.com/product/bug-bounty-platform
https://www.hackerone.com/product/response-vulnerability-disclosure-program
https://www.hackerone.com/product/pentest
https://www.hackerone.com/product/overview
https://hackerone.com/leaderboard
https://www.hackerone.com/resources/customer-story/how-hired-builds-customer-trust-with-hackerone-pentest
https://www.hackerone.com/customer-hub/Hyatt
https://www.hackerone.com/node/10766
https://hackerone.com/assets/static/css/vendor.8b11831d.css
https://hackerone.com/assets/static/css/main.b2076af6.css
```

## GoCrawler Go Library
```golang
import (
    types "github.com/m1dugh/gocrawler/pkg/types"
    "github.com/m1dugh/gocrawler/pkg/gocrawler"
)

func main() {
    includes := []string{`hackerone\.com`}
    excludes := []string{`api\.hackerone\.com`, `^http://`}
    scope := types.NewScope(includes, excludes)

    config := &gocrawler.Config{
        MaxThreads: 10,     // Sets the number of threads to 10
        MaxRequests: 10,    // Sets the max request rate to 10 per second
    }

    crawler := gocrawler.New(scope, config)
    urls := []string{`https://www.hackerone.com`}

    // Adds a callback when a url is requested.
    //  crawler.AddCallback(func(res *http.Response, body string) {
    //      fmt.Printf("url: %s, response length: %d\n", res.Request.URL, len(body))
    //  })

    crawler.Crawl(urls)
    for _, url := range crawler.Discovered.ToArray() {
        fmt.Println(url)
    }
}
```

## License
This project is built under the `GNU GPLv3 Protection License`.

