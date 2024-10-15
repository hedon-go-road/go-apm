#!/bin/bash

ab -n 6000 -c 100 "http://127.0.0.1:30001/order/add?uid=1&sku_id=3&num=1"