FROM postgres:latest

COPY ./wordlist.csv /home/wordlist.csv
COPY ./init.sql /docker-entrypoint-initdb.d/init.sql
