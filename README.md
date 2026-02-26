## GO Lang Microservices API Demo

#### To start everything, run the following command in your terminal:
>docker-compose up --build

#### To stop and clear any stale volumes:
>docker-compose down -v

Once the containers are running, you can interact with your microservices at the
following endpoints:

- `GET /health`
- `GET /select/{table}?where={condition}`
- `POST /insert/{table}`

#### CLI helper

>alias cli='./cli.sh'


#### Now can use it as
>cli [health | select | insert]