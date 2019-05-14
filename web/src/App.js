import React from 'react'
import { CopyToClipboard } from 'react-copy-to-clipboard'

const checkURLValidity = (inputURL) => {
  var validURLExpression = /(https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|www\.[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9]+\.[^\s]{2,}|www\.[a-zA-Z0-9]+\.[^\s]{2,})/gi
  var urlRegex = new RegExp(validURLExpression)
  return !inputURL.match(urlRegex)
}

const checkShortURLValidity = (inputURL) => {
  var urlRegex = new RegExp('^[a-z-]+$')
  return !urlRegex.test(inputURL)
}

export default class App extends React.Component {
  state = {
    longURLErrors: '',
    longURL: '',
    shortURLErrors: '',
    shortURL: '',
    isWorking: false,
    isCompleted: false,
  }

  async checkAndShortURL(longURL, shortURL) {
    const body = JSON.stringify({
      destinationURL: longURL.trim(),
      shortURL: shortURL.trim(),
    })

    await this.setState({
      isWorking: true,
    })

    try {
      const request = await fetch('/api/add', {
        headers: {
          'Content-Type': 'application/json',
        },
        method: 'POST',
        mode: 'cors',
        body,
      })
      const status = await request.status
      if (status === 201) {
        await this.setState({
          isWorking: false,
          isCompleted: true,
        })
      } else {
        await this.setState({
          isWorking: false,
          isCompleted: false,
        })
      }
    } catch (err) {
      console.error(err)
    }
  }

  async generateURL(e) {
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

  startNew(e) {
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

  handleFormChange(e) {
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

  render() {
    return (
      <main>
        <header className="App-header">AtomURL</header>
        <div className="container main-section">
          <div className="box-card shadow-lg rounded bg-white p-5">
            <form>
              <div className="input-group mb-3 input-group-lg">
                <input
                  name="longURL"
                  type="url"
                  className={`form-control ${
                    this.state.longURLErrors === '' ? '' : 'is-invalid'
                  } ${this.state.isCompleted ? 'form-control-plaintext' : ''}`}
                  aria-label="Sizing example input"
                  aria-describedby="inputGroup-sizing-default"
                  onChange={(e) => this.handleFormChange(e)}
                  autoFocus={true}
                  required={true}
                  placeholder="Destination URL"
                  disabled={this.state.isWorking || this.state.isCompleted}
                  value={this.state.longURL}
                />
                <div className="invalid-feedback">
                  {this.state.longURLErrors}
                </div>
              </div>
              <div className="text-center">
                <i className="fas fa-arrow-down h3 my-2" />
              </div>
              <div className="input-group input-group-lg mt-3 mb-5">
                {!this.state.isCompleted && (
                  <div className="input-group-prepend">
                    <span
                      className="input-group-text"
                      id="inputGroup-sizing-lg">
                      atomurl.ga/go/
                    </span>
                  </div>
                )}
                {this.state.isCompleted ? (
                  <>
                    <input
                      id="shortURLFinal"
                      type="text"
                      name="shortURL"
                      className={`form-control form-control-plaintext`}
                      aria-label="Sizing example input"
                      aria-describedby="inputGroup-sizing-lg"
                      value={`atomurl.ga/go/${this.state.shortURL}`}
                      required
                      disabled
                    />
                    <div class="input-group-append">
                      <CopyToClipboard
                        text={`atomurl.ga/go/${this.state.shortURL}`}>
                        <button
                          class="btn btn-info"
                          type="button"
                          id="button-addon2">
                          <i class="far fa-copy" />
                        </button>
                      </CopyToClipboard>
                    </div>
                  </>
                ) : (
                  <>
                    <input
                      type="text"
                      name="shortURL"
                      className={`form-control ${
                        this.state.shortURLErrors === '' ? '' : 'is-invalid'
                      }`}
                      aria-label="Sizing example input"
                      aria-describedby="inputGroup-sizing-lg"
                      onChange={(e) => this.handleFormChange(e)}
                      value={this.state.shortURL}
                      required
                      placeholder="Short URL"
                      disabled={this.state.isWorking}
                    />
                    <div className="invalid-feedback">
                      {this.state.shortURLErrors}
                    </div>
                  </>
                )}
              </div>
              <div className="text-center">
                {this.state.isCompleted ? (
                  <button
                    type="button"
                    className="btn btn-outline-info btn-lg mr-3"
                    onClick={(e) => this.startNew(e)}>
                    Successfully created, Add more ?
                  </button>
                ) : (
                  <button
                    type="submit"
                    className="btn btn-info btn-lg"
                    onClick={(e) => this.generateURL(e)}
                    disabled={this.state.isWorking}>
                    <div className="d-flex">
                      {this.state.isWorking
                        ? 'Checking and adding URL'
                        : 'Assign the above URL'}
                      <span
                        class={`spinner-grow spinner-grow-md ml-3 fade animate-all ${
                          this.state.isWorking ? '' : 'd-none'
                        }`}
                        role="status"
                        aria-hidden="true"
                      />
                    </div>
                  </button>
                )}
              </div>
            </form>
          </div>
        </div>
      </main>
    )
  }
}
