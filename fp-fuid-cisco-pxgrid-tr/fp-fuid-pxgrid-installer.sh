#!/bin/bash

chmod +x fuid-ise
mkdir /var/fuid-ise
mkdir /var/fuid-ise/fuid-ise-logs
mkdir /var/fuid-ise/latest-timestamp
mv fuid-ise.service /etc/systemd/system/
mv fuid-ise /var/fuid-ise/
mv fuid-ise.yml /var/fuid-ise/
sudo systemctl enable fuid-ise.service