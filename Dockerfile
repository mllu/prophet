FROM golang:1.11

WORKDIR /app

# Set an env var that matches your github repo name, replace treeder/dockergo here with your repo name
ENV SRC_DIR=/go/src/github.robot.car/meng-lin-lu/prophet

# Add the source code:
ADD . /app/src/prophet

# Build it:
RUN cd /app/src/prophet; go build -o prophet; chmod +x prophet; cp prophet /app/

# expost port
EXPOSE 8080

ENTRYPOINT ["./prophet"]

