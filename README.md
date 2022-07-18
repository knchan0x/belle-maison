# BM Tracker

A price tracker of Belle Maison Products.

## Features

1. Notice the user when the price of traced product hits the target price

2. Notice user when the stock of traced product is less than 10

## Built by

Dashboard: vue3

Backend: gin, goquery, gocron, viper, gorm, mysql, docker

## Usage

1. Rename config_template.yaml to config.yaml and edit it.
2. Rename docker-compose_template.yml to docker-compose.yml and edit it.
3. run ```cd ./dashboard && npm run build``` if you have amended the dashboard and then run ```cd ../``` back to root folder.
4. run ```cd ./backend/cmd/web && go run .``` for testing locally.
5. open http://localhost/bellemaison/ to view.
6. run ```docker-compose build``` to build docker image.
7. upload it and open https://www.yoursite.com/bellemaison

## TODO

### New Functions
1. Disable notification for specific product
2. Modify target price
3. Modify recipients' email address in dashboard

# Licence
MIT