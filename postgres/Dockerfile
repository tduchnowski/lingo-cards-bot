FROM postgres:latest

# copy dictionaries, frequencies and wiktionary dumps to container
COPY dictionaries/ /docker-entrypoint-initdb.d/dictionaries
COPY frequencies/ /docker-entrypoint-initdb.d/frequencies
COPY wiktionary/ /docker-entrypoint-initdb.d/wiktionary

COPY ./init.sql /docker-entrypoint-initdb.d/init.sql
