// productservice.js
import callApi from './api';

export const getAllProducts = async (filters = {}) => {
  const params = new URLSearchParams(filters).toString();
  return callApi("product", `/product?${params}`);
};

export const getProductByEAN = async (ean) => {
  const products = await callApi("product", `/product?eans=${ean}`);
  return products && products.length > 0 ? products[0] : null;
};

export const addReview = async (reviewData) => {
  return callApi("product", `/review`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(reviewData),
  });
};


