FROM fedora:38

WORKDIR /app

RUN dnf -y install golang nodejs vips-devel

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY package.json package-lock.json ./
RUN npm install

CMD ["/root/go/bin/air"]