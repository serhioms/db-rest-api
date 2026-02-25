To start everything, run the following command in your terminal:
docker-compose up --build

Once the containers are running, you can interact with your microservices at the
following endpoints:

* GET /health
* GET /select/{table}?where={condition}
* POST /insert/{table}

  You can now run the project again:
  1 docker-compose down -v  # Clear any stale volumes
  2 docker-compose up --build