import React, { useState, useEffect } from 'react';
// Zorg ervoor dat 'useHistory' de juiste import is voor jouw React Router versie (v5)
import { useHistory } from 'react-router-dom';
import { List, Card, Spin, message, Input, Tag } from 'antd'; // Tag toegevoegd voor tags
import Container from '../common/Container'; // Zorg dat dit pad klopt
import { getAllProducts } from '../services/productService'; // Zorg dat dit pad klopt

const { Search } = Input;

const ProductListPage = () => {
    const history = useHistory(); // Correct geïnitialiseerd
    const [products, setProducts] = useState([]);
    const [loading, setLoading] = useState(false);
    const [searchTerm, setSearchTerm] = useState(''); // Voor de zoekbalk
    const [filteredProducts, setFilteredProducts] = useState([]); // Nieuwe state voor gefilterde producten

    useEffect(() => {
        fetchProducts();
    }, []);

    // Effect voor het filteren van producten op basis van zoekterm
    useEffect(() => {
        if (searchTerm) {
            const lowerCaseSearchTerm = searchTerm.toLowerCase();
            const filtered = products.filter(product =>
                product.naam.toLowerCase().includes(lowerCaseSearchTerm) ||
                product.ean.toString().includes(lowerCaseSearchTerm) ||
                (product.omschrijving && product.omschrijving.toLowerCase().includes(lowerCaseSearchTerm)) ||
                (product.merk && product.merk.toLowerCase().includes(lowerCaseSearchTerm)) ||
                (product.categorieen && product.categorieen.some(cat => cat.naam.toLowerCase().includes(lowerCaseSearchTerm))) ||
                (product.tags && product.tags.some(tag => tag.naam.toLowerCase().includes(lowerCaseSearchTerm)))
            );
            setFilteredProducts(filtered);
        } else {
            setFilteredProducts(products); // Als zoekterm leeg is, toon alle producten
        }
    }, [searchTerm, products]); // Filter opnieuw wanneer zoekterm of producten veranderen


    const fetchProducts = async (filters = {}) => {
        setLoading(true);
        try {
            const fetchedProducts = await getAllProducts(filters);
            setProducts(fetchedProducts || []);
        } catch (error) {
            message.error("Fout bij het laden van producten.");
            console.error("Error fetching products:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleProductClick = (ean) => {
        history.push(`/products/${ean}`);
    };

    const onSearch = (value) => {
        setSearchTerm(value);
    };

    // Functie om de laagste prijs te vinden
    const getLowestPrice = (productAanbod) => {
        if (!productAanbod || productAanbod.length === 0) {
            return 'N/A';
        }
        const prices = productAanbod.map(offer => offer.prijs);
        return Math.min(...prices).toFixed(2);
    };

    if (loading) {
        return (
            <Container>
                <Spin size="large" tip="Producten laden..." />
            </Container>
        );
    }

    return (
        <Container>
            <h1>Alle Producten</h1>
            <Search
                placeholder="Zoek producten op naam, EAN, omschrijving, merk, categorie of tag..."
                onSearch={onSearch}
                onChange={(e) => setSearchTerm(e.target.value)}
                style={{ width: '100%', marginBottom: '20px' }}
                value={searchTerm} // Zorgt dat de input box de state reflecteert
            />
            {filteredProducts.length > 0 ? (
                <List
                    grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 4, xl: 4, xxl: 4 }}
                    dataSource={filteredProducts} // Gebruik de gefilterde producten
                    renderItem={(product) => (
                        // Gebruik de camelCase keys die de backend teruggeeft
                        <List.Item onClick={() => handleProductClick(product.ean)}>
                            <Card
                                hoverable
                                // Gebruik de eerste afbeelding uit de 'afbeeldingen' array
                                cover={
                                    product.afbeeldingen && product.afbeeldingen.length > 0 ? (
                                        <img
                                            alt={product.naam || 'Product afbeelding'}
                                            src={product.afbeeldingen[0]}
                                            style={{ height: '200px', objectFit: 'cover', width: '100%' }}
                                        />
                                    ) : (
                                        <div style={{ height: '200px', background: '#f0f0f0', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#888' }}>
                                            Geen afbeelding
                                        </div>
                                    )
                                }
                            >
                                <Card.Meta
                                    title={product.naam || 'Onbekend Product'}
                                    description={
                                        <>
                                            <p>EAN: {product.ean}</p>
                                            <p>Merk: {product.merk || 'N/A'}</p>
                                            {/* Toon de laagste prijs */}
                                            <p>Prijs: &euro;{getLowestPrice(product.productAanbod)}</p>
                                            <p>{product.omschrijving || 'Geen beschrijving'}</p>
                                            {/* Categorieën tonen */}
                                            {product.categorieen && product.categorieen.length > 0 && (
                                                <p>
                                                    <strong>Categorieën: </strong>
                                                    {product.categorieen.map((cat, index) => (
                                                        <Tag key={cat.ID || index}>{cat.naam}</Tag>
                                                    ))}
                                                </p>
                                            )}
                                            {/* Tags tonen */}
                                            {product.tags && product.tags.length > 0 && (
                                                <p>
                                                    <strong>Tags: </strong>
                                                    {product.tags.map((tag, index) => (
                                                        <Tag key={tag.id || index}>{tag.naam}</Tag>
                                                    ))}
                                                </p>
                                            )}
                                        </>
                                    }
                                />
                            </Card>
                        </List.Item>
                    )}
                />
            ) : (
                <p>Geen producten gevonden die voldoen aan de zoekcriteria.</p>
            )}
        </Container>
    );
};

export default ProductListPage;