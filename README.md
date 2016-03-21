# hbp-performance-test

Setup
===============

1. Install golang from here: https://golang.org/dl/
2. Configure the script to your needs: see Config
3. go run httpRequest.go (runs the script on command line)
 
Config
---------------

You should set the hostname at the very least, including the port if not standard http. Then set the number ob parallel bunches. The bunch size is currently 8, but can also be set.

    const numberOfBunches = 10
    const bunchSize int = 8

The image path is currently set as follows:

    const imagePath = "/srv/data/HBP/BigBrain_jpeg.h5"

The script currently does not allow to use multiple images as the metadata is not yet read automatically. The requestParameterDictionary is currently statically defined and just to be used with the standard image "/srv/data/HBP/BigBrain_jpeg.h5". If other images' performance shall be measured one has to adopt the requestParameterDictionary accordingly or read json through the metadata URL and pack the returned json into the requestParameterDictionary. This is work for the future if needed.

Scaling
---------------
This script should scale to multiple thousands of concurrent requests.
