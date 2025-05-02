import { useState } from 'react';
import axios from 'axios';
import { useCart } from '../context/CartContext';

function Payments() {
  const [cardNumber, setCardNumber] = useState('');
  const { cartId, cart, fetchCart } = useCart();
  const [message, setMessage] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await axios.post('http://localhost:1323/payments', {
        cart_id: cartId,
        card_number: cardNumber,
        amount: cart.reduce((sum, item) => sum + item.price, 0)
      });
      setMessage('Płatność zakończona sukcesem!');
      fetchCart([]);
    } catch (error) {
      setMessage('Błąd płatności');
    }
  };

  return (
    <div>
      <h2>Płatności</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Numer karty"
          value={cardNumber}
          onChange={(e) => setCardNumber(e.target.value)}
          required
        />
        <button type="submit">Zapłać</button>
      </form>
      {message && <p>{message}</p>}
    </div>
  );
}

export default Payments;