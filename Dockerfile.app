FROM scratch

COPY piggy piggy

CMD ["./piggy", "start"]
