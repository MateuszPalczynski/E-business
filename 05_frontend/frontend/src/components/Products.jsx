import { useState, useEffect } from 'react';
import axios from 'axios';
import { useCart } from '../context/CartContext';

function Products() {
  const [products, setProducts] = useState([]);
  const { addToCart } = useCart();

  useEffect(() => {
    const fetchProducts = async () => {
      const response = await axios.get('http://localhost:1323/products');
      setProducts(response.data);
    };
    fetchProducts();
  }, []);

  return (
    <div>
      <h2>Produkty</h2>
      <div className="products-list">
        {products.map(product => (
          <div key={product.id} className="product-card">
            <h3>{product.name}</h3>
            <p>Cena: {product.price} zł</p>
            <button onClick={() => addToCart(product.id)}>Dodaj do koszyka</button>
          </div>
        ))}
      </div>
    </div>
  );
}

export default Products;