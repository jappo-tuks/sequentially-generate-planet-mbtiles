package downloadsubregion

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type DownloadInformation struct {
	SubRegion       string
	ContentLength   int
	TotalDownloaded int
}

func (di *DownloadInformation) Write(p []byte) (n int, err error) {
	di.TotalDownloaded += len(p)
	percentage := di.TotalDownloaded * 100 / di.ContentLength
	log.Printf("Downloaded %v of %v bytes (%v%%) of %v\n", di.TotalDownloaded, di.ContentLength, percentage, di.SubRegion)
	return di.TotalDownloaded, nil
}

func DownloadSubRegion(subRegion, destFolder string) (ok bool, err error) {
	subRegionUrl := "https://download.geofabrik.de/" + subRegion + "-latest.osm.pbf"

	// HEAD request - find out the content-length (file size in bytes).
	r, err := http.NewRequest("HEAD", subRegionUrl, nil)
	if err != nil {
		return false, err
	}
	rH, err := http.DefaultClient.Do(r)
	if err != nil {
		return false, err
	}
	contentLength, err := strconv.Atoi(rH.Header.Get("Content-Length"))
	if err != nil {
		log.Printf("No content length was provided for the download, therefore progress cannot be displayed. The file will still be downloaded.\n")
	}
	fmt.Printf("Content Length: %T, %v\n", contentLength, contentLength)

	// Setup file to write download to.
	fileName := strings.Split(subRegion, "/")
	f, err := os.Create(destFolder + "/" + fileName[len(fileName)-1] + ".osm.pbf")
	if err != nil {
		log.Printf("Failed to create %v.osm.pbf file, in %v\n", subRegion, destFolder)
		return false, err
	}

	// GET request
	r.Method = "GET"
	rG, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer rG.Body.Close()

	// Initiate struct implementing writer for progress reporting.
	di := &DownloadInformation{
		SubRegion:       subRegion,
		ContentLength:   contentLength,
		TotalDownloaded: 0,
	}

	// Write to file and progress report.
	tee := io.TeeReader(rG.Body, di)
	_, err = io.Copy(f, tee)
	if err != nil {
		log.Printf("There was a problem writing to file: %v.osm.pbf\n", subRegion)
		return false, err
	}
	return true, nil
}
