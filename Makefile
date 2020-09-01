.PHONY: test build dist release release-dryrun clean

NAME = apicmp

DIST_OPTS = -a -tags netgo -installsuffix netgo
BUILD_CMD = go build $(DIST_OPTS)
SRC_FILES = main.go
DIST_DIR = ./dist

test:
	go test -cover ./...

build:
	go build -o $(NAME) $(SRC_FILES)

$(NAME)-%:
	GOOS=$* $(BUILD_CMD) -o $(DIST_DIR)/$@ $(SRC_FILES)

dist: $(NAME)-darwin $(NAME)-linux $(NAME)-windows

release: dist
	semantic-release -noci -ghr -vf -slug arithran/apicmp
	ghr $$(cat .ghr) dist

release-dryrun:
	semantic-release -noci -ghr -vf -slug arithran/apicmp -dry

clean:
	rm -f $(NAME)
	rm -rf $(DIST_DIR)
