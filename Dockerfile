FROM golang:1.11

WORKDIR /app

# Set an env var that matches your github repo name, replace treeder/dockergo here with your repo name
ENV SRC_DIR=/go/src/github.robot.car/meng-lin-lu/prophet

# Add the source code:
ADD . $SRC_DIR

# Build it:
RUN cd $SRC_DIR; go build -o prophet; chmod +x prophet; cp prophet /app/

# expost port
EXPOSE 8080

ENTRYPOINT ["./prophet"]

