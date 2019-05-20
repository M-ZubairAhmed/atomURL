import React from 'react'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { createBrowserHistory } from 'history'
import logo from './assets/favicon.png'
import { ReactComponent as DownArrowSVG } from './assets/arrow-down-solid.svg'
import { ReactComponent as CopySVG } from './assets/copy-solid.svg'

const PAGE_INIT = 'init'
const PAGE_NO_FOUND = '404'
const PAGE_HOME = 'home'

const checkURLValidity = (inputURL) => {
  var validURLExpression = /(https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|www\.[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9]+\.[^\s]{2,}|www\.[a-zA-Z0-9]+\.[^\s]{2,})/gi
  var urlRegex = new RegExp(validURLExpression)
  return !inputURL.match(urlRegex)
}

const checkShortURLValidity = (inputURL) => {
  var urlRegex = new RegExp('^[a-z-]+$')
  return !urlRegex.test(inputURL)
}

const DestinationURLHolder = ({
  url,
  errors = '',
  isCompleted = false,
  isWorking = false,
  handleFormChange,
}) => {
  return (
    <div className="form-group">
      <div className="mb-3 input-group-lg">
        <label for="input_longURL font-weight-lighter">Destination URL</label>
        <input
          name="longURL"
          id="input_longURL"
          type="url"
          className={`form-control ${errors === '' ? '' : 'is-invalid'} ${
            isCompleted ? 'form-control-plaintext' : ''
          }`}
          aria-label="Destination url"
          aria-describedby="Enter the url which you want to redirect to"
          autoFocus={true}
          required={true}
          title="Enter the url which you want to redirect to"
          disabled={isWorking || isCompleted}
          onChange={(e) => handleFormChange(e)}
          value={url}
        />
        <div className="invalid-feedback">{errors}</div>
      </div>
    </div>
  )
}

const DownArrow = () => (
  <div className="down-arrow">
    <DownArrowSVG />
  </div>
)

const ShortURLHolder = ({
  url,
  errors = '',
  isCompleted = false,
  isWorking = false,
  handleFormChange,
}) => {
  // Form page editing
  if (isCompleted) {
    return (
      <div className="input-group input-group-lg mt-3 mb-5">
        <input
          id="shortURLFinal"
          type="text"
          name="shortURL"
          className={`form-control form-control-plaintext`}
          aria-label="Short url"
          aria-describedby="Short url you assigned"
          required
          disabled
          value={`atomurl.ga/go/${url}`}
        />
        <div className="input-group-append">
          <CopyToClipboard text={`atomurl.ga/go/${url}`}>
            <button
              className="btn btn-info"
              type="button"
              id="button-addon2"
              title="Copy to clipboard">
              <div className="copy-icon  ">
                <CopySVG />
              </div>
            </button>
          </CopyToClipboard>
        </div>
      </div>
    )
  } else {
    return (
      <>
        <div className="d-block d-md-none text-muted font-weight-light font-italic">{`atomurl.ga/go/${url}`}</div>
        <div className="input-group input-group-lg mt-3 mb-5">
          <div className="input-group-prepend d-none d-md-block">
            <span
              className="input-group-text font-weight-bold"
              id="inputGroup-sizing-lg">
              atomurl.ga/go/
            </span>
          </div>
          <input
            id="shortURLFinal"
            type="text"
            name="shortURL"
            className={`form-control ${errors === '' ? '' : 'is-invalid'}`}
            aria-label="Short url"
            aria-describedby="Enter short url which you wish to assign"
            title="Enter short url which you wish to assign"
            required
            disabled={isWorking}
            onChange={(e) => handleFormChange(e)}
            value={url}
          />
          <div className="invalid-feedback">{errors}</div>
        </div>
      </>
    )
  }
}

const ActionButton = ({
  isWorking = false,
  isCompleted = false,
  startNewForm,
  linkBothUrls,
}) => {
  if (isCompleted) {
    return (
      <div className="text-center">
        <button
          type="button"
          className="btn btn-outline-info btn-lg mr-3"
          onClick={(e) => startNewForm(e)}>
          Successfully linked, Add more ?
        </button>
      </div>
    )
  } else {
    return (
      <div className="text-center">
        <button
          type="submit"
          className="btn btn-info btn-lg"
          onClick={(e) => linkBothUrls(e)}
          disabled={isWorking}>
          <div className="d-flex">
            {isWorking
              ? 'Checking and adding url'
              : 'Link both of the above urls'}
            <span
              className={`spinner-grow spinner-grow-md ml-3 fade animate-all ${
                isWorking ? '' : 'd-none'
              }`}
              role="status"
              aria-hidden="true"
            />
          </div>
        </button>
      </div>
    )
  }
}

export default class App extends React.Component {
  state = {
    longURLErrors: '',
    longURL: '',
    shortURLErrors: '',
    shortURL: '',
    isWorking: false,
    isCompleted: false,
    page: PAGE_INIT,
  }

  history = createBrowserHistory()

  redirectToHome(e) {
    e.preventDefault()
    this.history.replace('/')
    this.setState({ page: PAGE_HOME })
  }

  async checkAndShortURL(longURL, shortURL) {
    const body = JSON.stringify({
      destinationURL: longURL.trim(),
      shortURL: shortURL.trim(),
    })

    await this.setState({
      isWorking: true,
    })

    const ADD_SHORT_URL =
      process.env.NODE_ENV === 'production'
        ? '/api/add'
        : 'http://localhost:8000/api/add'

    try {
      const request = await fetch(ADD_SHORT_URL, {
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'POST',
        mode: 'cors',
        body,
      })
      const status = await request.status
      const response = await request.json()

      if (response) {
        if (status === 201) {
          await this.setState({
            isWorking: false,
            isCompleted: true,
          })
        } else {
          // eslint-disable-next-line
          throw { ...response, status }
        }
      } else {
        // eslint-disable-next-line
        throw 'UNKNOW_ERROR'
      }
    } catch (err) {
      let shortURLErrors = ''
      if (err && err.status) {
        if (err.status === 409) {
          shortURLErrors = 'Short url already take, please try something else'
        } else {
          shortURLErrors = err.error
        }
      } else {
        shortURLErrors =
          'Something terrible happened while we tried linking the url, try again later'
      }
      await this.setState({
        isWorking: false,
        isCompleted: false,
        shortURLErrors,
      })
      console.error('Error in linking url', err)
    }
  }

  linkBothUrls = async (e) => {
    e.preventDefault()

    const { longURL, shortURL } = this.state

    let longURLErrors = ''
    let shortURLErrors = ''

    if (checkURLValidity(longURL)) {
      longURLErrors = 'URL should be of format <http>://domain.<tld>'
    }
    if (shortURL.length === 0) {
      shortURLErrors = 'This cannot be empty'
    } else if (checkShortURLValidity(shortURL)) {
      shortURLErrors =
        'Shortened url can only contain small letters and hyphens ( - )'
    } else if (
      shortURL.trim().slice(0, 1) === '-' ||
      shortURL.trim().slice(-1) === '-'
    ) {
      shortURLErrors = 'It cannot start or end with hyphen ( - )'
    }

    // Settings state for errors
    if (longURLErrors || shortURLErrors) {
      this.setState({
        longURLErrors,
        shortURLErrors,
      })
    } else {
      // Frontend no errors
      this.checkAndShortURL(longURL, shortURL)
    }
  }

  startNewForm = (e) => {
    e.preventDefault()
    this.setState({
      isWorking: false,
      isCompleted: false,
      longURL: '',
      shortURL: '',
      longURLErrors: '',
      shortURLErrors: '',
    })
  }

  handleFormChange = (e) => {
    let validation = ''
    const {
      target: { name, value },
    } = e

    if (name === 'longURL') {
      validation = 'longURLErrors'
    } else if (name === 'shortURL') {
      validation = 'shortURLErrors'
    }

    this.setState({
      [name]: value,
      [validation]: '',
    })
  }

  componentDidMount() {
    const {
      location: { pathname },
    } = this.history

    if (pathname === '/404') {
      this.setState({ page: PAGE_NO_FOUND })
    } else {
      this.setState({ page: PAGE_HOME })
    }
  }

  render() {
    const {
      state: {
        isCompleted,
        isWorking,
        longURL,
        longURLErrors,
        shortURL,
        shortURLErrors,
        page,
      },
      handleFormChange,
      linkBothUrls,
      startNewForm,
    } = this

    return (
      <main>
        <header className="App-header">
          <img src={logo} alt="AtomURL" />
        </header>
        <div className="container mt-5">
          <div className="row justify-content-center">
            <div className="col col-md-6">
              {page === PAGE_INIT && (
                <div className="text-center mt-5">
                  <div className="lds-heart text-center mt-4">
                    <div />
                  </div>
                  <div className="lead">Loading !!</div>
                </div>
              )}
              {page === PAGE_HOME && (
                <div className="box-card shadow-lg rounded bg-white p-3 p-md-5">
                  <form>
                    <DestinationURLHolder
                      url={longURL}
                      handleFormChange={handleFormChange}
                      errors={longURLErrors}
                      isWorking={isWorking}
                      isCompleted={isCompleted}
                    />
                    <DownArrow />
                    <ShortURLHolder
                      url={shortURL}
                      handleFormChange={handleFormChange}
                      errors={shortURLErrors}
                      isWorking={isWorking}
                      isCompleted={isCompleted}
                    />
                    <ActionButton
                      isCompleted={isCompleted}
                      isWorking={isWorking}
                      linkBothUrls={linkBothUrls}
                      startNewForm={startNewForm}
                    />
                  </form>
                </div>
              )}
              {page === PAGE_NO_FOUND && (
                <div className="text-center mt-5">
                  <h2>Opps! URL entered doesnt redirect anywhere</h2>
                  <button
                    className="btn btn-info btn-lg mt-5"
                    onClick={(e) => this.redirectToHome(e)}>
                    Lets create a new short url
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
    )
  }
}
