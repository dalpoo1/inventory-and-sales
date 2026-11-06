import { NavLink, Route, Routes } from 'react-router-dom'
import OrderDetailPage from './pages/OrderDetailPage'
import OrdersPage from './pages/OrdersPage'
import ProductsPage from './pages/ProductsPage'

export default function App() {
  const navClass = ({ isActive }: { isActive: boolean }) =>
    `rounded-md px-3 py-1.5 text-sm font-medium ${
      isActive ? 'bg-blue-50 text-blue-700' : 'text-gray-600 hover:bg-gray-50'
    }`

  return (
    <div className="min-h-screen bg-gray-100 text-gray-900">
      <nav className="border-b border-gray-200 bg-white">
        <div className="mx-auto flex max-w-6xl items-center gap-6 px-6 py-3">
          <span className="text-lg font-bold text-gray-900">StockFlow</span>
          <div className="flex gap-1">
            <NavLink to="/" end className={navClass}>
              Products
            </NavLink>
            <NavLink to="/orders" className={navClass}>
              Orders
            </NavLink>
          </div>
        </div>
      </nav>

      <main className="mx-auto max-w-6xl px-6 py-6">
        <Routes>
          <Route path="/" element={<ProductsPage />} />
          <Route path="/orders" element={<OrdersPage />} />
          <Route path="/orders/:id" element={<OrderDetailPage />} />
        </Routes>
      </main>
    </div>
  )
}
