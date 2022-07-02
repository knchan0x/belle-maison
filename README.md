# BM Tracker

A price tracker of Belle Maison Products.

## Features

1. Notice the user when the price of traced product hits the target price

2. Notice user when the stock of traced product is less than 10

## Built by

Dashboard: vue3

Backend: gin, goquery, gocron, viper, gorm, mysql, docker

## Usage

1. run ```npm run build``` if you have amended the dashboard.
2. Rename config_template.yaml config.yaml and edit it.
3. Rename docker-compose_template.yml to docker-compose.yml and edit it.
4. run ```docker-compose build```.
5. open www.yoursite.com/bellemaison

# Licence
MIT