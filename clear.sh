#!/bin/bash

ps -ef | grep "headless" | cut -d " " -f4 |xargs kill -9
