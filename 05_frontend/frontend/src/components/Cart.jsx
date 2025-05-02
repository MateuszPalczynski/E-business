import { useState, useEffect } from 'react';
import { useCart } from '../context/CartContext';
import { Link } from 'react-router-dom';

const Cart = () => {
  const { cart, fetchCart, removeFromCart } = useCart();
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetchCart();
    setTotal(cart.reduce((sum, item) => sum + item.price, 0));
  }, [cart, fetchCart]);

  return (
    <div className="cart-container">
      <h2>Twój Koszyk</h2>
      {cart.length === 0 ? (
        <p className="empty-cart">Koszyk jest pusty</p>
      ) : (
        <>
          <div className="cart-items">
            {cart.map(item => (
              <div key={item.id} className="cart-item">
                <div className="item-info">
                  <h3>{item.name}</h3>
                  <p>Cena: {item.price.toFixed(2)} zł</p>
                </div>
                <button 
                  className="remove-button"
                  onClick={() => removeFromCart(item.id)}
                >
                  ×
                </button>
              </div>
            ))}
          </div>
          <div className="cart-summary">
            <h3>Suma całkowita: {total.toFixed(2)} zł</h3>
            <Link to="/payments" className="checkout-button">
              Przejdź do płatności
            </Link>
          </div>
        </>
      )}
    </div>
  );
}

export default Cart;