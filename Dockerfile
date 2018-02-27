# our base image
FROM ubuntu
ARG GIT_COMMIT=unkown
LABEL git-commit=$GIT_COMMIT
# copy the build file
#RUN ["mkdir", "/book_server"]
COPY . /book_server
WORKDIR /book_server

# specify the port number the container should expose
EXPOSE 10000

# run the server
# CMD ["cd", "/book_server"]
CMD ["./main"]