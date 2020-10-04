FROM scratch

VOLUME /database

COPY piggy piggy

CMD ["./piggy", "start"]
