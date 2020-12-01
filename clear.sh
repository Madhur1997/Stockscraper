#!/bin/bash

# Clear out the running headless Chrome instances if you kill the program without chromedp getting a chance to clear up.
ps -ef | grep "headless" | cut -d " " -f4 |xargs kill -9
