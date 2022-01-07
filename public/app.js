class App {
  constructor(window) {
    showVersion(window)
    defineCustomElements(window)
    if (loggedIn(window)) {
      console.info('Logged in, refreshing app data.')
    }
    else {
      console.debug('Not logged in, showing login options.')
      showLoginOptions(window)
    }
  }
}

const loggedIn = ({ localStorage, console, JSON }) => {
  const json = localStorage.session
  if (json) {
    try {
      const { token, expiresAt } = JSON.parse(json)
      return true
    }
    catch(error) {
      console.debug(error)
      console.warn(`session invalid, deleting from localStorage`)
      delete localStorage.session
      return false
    }
  }
}

const showLoginOptions = async ({ document }) => {
  const links = document.createElement('login-links')
  forEach(document.querySelectorAll('main'), replaceChildren(links))
}

const replaceChildren = (children) => {
  return (parent) => parent.replaceChildren(children)
}

const template = ({ document }, id) => {
  return document.querySelector(`template#${id}`)?.content?.cloneNode(true)
}

const forEach = (iterable, f) => { [...iterable].forEach(f) }

const showVersion = async ({ document, fetch, console }) => {
  const response = await fetch('version.txt')
  const text = await response.text()
  const version = document.createTextNode(text)
  const sourceLink = document.querySelector('link#source[rel="help"]')
  const href = `${sourceLink?.href}/commit/${version}`
  forEach(document.querySelectorAll('a#commit'), (e) => {
    e.replaceChildren(version)
    if (sourceLink) e.setAttribute('href', href)
  })
}

const defineCustomElements = ({ document, customElements, fetch }) => {
  customElements.define('login-links', class extends HTMLElement {
    constructor() {
      super()
      const ul = document.createElement('ul')
      fetch('/oauth2/providers').then(r => r.json()).then((providers) => {
        providers.forEach(({ id, title, href, icon }) => {
          const li = document.createElement('li')
          const a = document.createElement('a')
          const img = document.createElement('img')
          img.setAttribute('src', icon)
          img.setAttribute('title', title)
          img.setAttribute('alt', id)
          a.setAttribute('href', href)
          a.addEventListener('click', this)
          a.appendChild(img)
          li.appendChild(a)
          ul.appendChild(li)
        })
        this.appendChild(ul)
      })
    }
    handleEvent(event) {
      event.preventDefault()
      const { target: { href } } = event
      const { location } = document
      location.assign(href)
    }
  })
}

export default App
