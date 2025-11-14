import { Link } from 'react-router-dom'
import './Header.css'

function Header() {
  return (
    <header className="header">
      <div className="header-container">
        <Link to="/" className="logo">
          <h1>Lingopaste</h1>
        </Link>
        <nav className="nav">
          <Link to="/">New Paste</Link>
          <Link to="/account">Account</Link>
        </nav>
      </div>
    </header>
  )
}

export default Header
