# hbp-performance-test

Setup
===============

1. Install golang from here: https://golang.org/dl/
2. Configure the script to your needs: see Config
3. go run httpRequest.go MetaData.go (runs the script on command line)
 
Config
---------------

You should set the hostname at the very least, including the port if not standard http. Then set the number of parallel bunches. The bunch size is currently 8, but can also be set.

    const numberOfBunches = 10
    const bunchSize int = 8

The image path is currently set as follows:

    const imagePath = "/srv/data/HBP/template/human/bigbrain_20um/sections/bigbrain.h5"

Currently the default is set to create static tile requests, which will create the very same tile requests for every single performance test run. There is also a random tile request mode, which can be switched on by

    const randomTileRequests = true

Metdata for the images is queried automatically. One word of caution: Getting metadata for some images did not work, e.g. const imagePath = "/srv/data/HBP/template/rat/waxholm/v2/anno/whs_axial_v2.h5". Please feel free to further investigate.

There are several ways to increase the performance related output (const reportDetail = AverageTimeOverBunches) and have debugging enabled (init function).

Scaling
---------------
This script should scale to multiple thousands of concurrent requests.
