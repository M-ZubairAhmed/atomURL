<p align="center">
  <img src="https://user-images.githubusercontent.com/17708702/57923341-aa22ba00-78bf-11e9-85fc-53a91a5577fc.png" alt="repo image" width="400" height="100" />
  <h3 align="center">https://atomurl.ga</h3>
  <h5 align="center"><i>A fast url shortener without trackers</i></h5>
</p>


### Project motivation

I wanted to make a simple url choose your own shortener without any trackers or adds, hence this project. Also to keep my learning going in world of Go.

### Starting locally

If you would like to host your own or start locally follow these steps :

- Clone the repository

```bash
git clone https://github.com/M-ZubairAhmed/atomURL.git
```

- Go to root folder and install go dependencies

```bash
go get -v ./...
```

- Go to `web/` folder and install react dependencies

```bash
cd web/
yarn
```

- Build the react project

```bash
yarn build
```

- Back to root folder and go through `.env.sample` file

```bash
cd ../
vi .env.sample
```

- Obtain environment values of mongoDB url either for local version or hosted version (eg.MongoDB cloud atlas)

```yml
# format of mongo database url
Host://UserName:Password@Address/databaseName
```

- Add obtained environmental values in `env.sample` and rename it.

```bash
# after adding values in file env.sample
mv .env.sample .env
```

- Replace the values in go run command

```bash
DB_HOST=insert_value_here DB_USER=insert_value_here DB_PASSWORD=insert_value_here DB_URL=insert_value_here DB_NAME=insert_value_here go run main.go
```

### Thanks goes to

- [Gophercises](https://gophercises.com/) by [Jon Calhoun](https://twitter.com/joncalhoun)
- [V Ramana](https://github.com/vramana)
- [Discord gophers](https://discordapp.com/invite/64C346U)
