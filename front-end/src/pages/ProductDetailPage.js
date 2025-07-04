import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Card, Spin, message, Typography, Divider, Form, Input, Button, Rate, List, Tag, Image, Carousel } from 'antd'; // Image, Carousel toegevoegd
import Container from '../common/Container'; // Zorg dat dit pad klopt
import { getProductByEAN, addReview } from '../services/productService'; // Zorg dat dit pad klopt en dat productservice.js de juiste velden voor addReview stuurt

const { Title, Paragraph, Text } = Typography; // Text toegevoegd
const { TextArea } = Input;

const ProductDetailPage = () => {
    const { ean } = useParams();
    const [product, setProduct] = useState(null);
    const [loading, setLoading] = useState(false);
    const [form] = Form.useForm(); // Ant Design form hook

    // Functie om de laagste prijs te vinden uit productAanbod
    const getLowestPrice = (productAanbod) => {
        if (!productAanbod || productAanbod.length === 0) {
            return 'N/A';
        }
        const prices = productAanbod.map(offer => offer.prijs);
        return Math.min(...prices).toFixed(2);
    };

    useEffect(() => {
        if (ean) {
            fetchProductDetails(ean);
        }
    }, [ean]);

    const fetchProductDetails = async (productEAN) => {
        setLoading(true);
        try {
            const fetchedProduct = await getProductByEAN(productEAN); // Haal het product op
            setProduct(fetchedProduct);
            // Je reviews zitten direct in het product object, dus we hoeven geen aparte state te vullen
            // setReviews(fetchedProduct.Reviews || []); // Deze lijn kan weg, reviews zijn direct toegankelijk via product.reviews
        } catch (error) {
            message.error("Fout bij het laden van productdetails.");
            console.error("Error fetching product details:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleAddReview = async (values) => {
        setLoading(true);
        try {
            const reviewData = {
                // Let op: Backend verwacht 'productEAN' (camelCase)
                productEAN: parseInt(ean), // Zorg dat EAN een nummer is
                naam: values.name, // Gebruik 'name' van het formulier
                score: values.score, // Gebruik 'score' van het formulier
                titel: values.title, // Gebruik 'title' van het formulier
                inhoud: values.content, // Gebruik 'content' van het formulier
                // gebruikerid: "gebruiker-id-van-jwt" // Als je dit nodig hebt van je JWT, moet je het hier toevoegen
            };
            await addReview(reviewData); // Roep je addReview service aan
            message.success("Review succesvol toegevoegd!");
            form.resetFields(); // Leeg het formulier
            fetchProductDetails(ean); // Herlaad productdetails om de nieuwe review te zien
        } catch (error) {
            message.error(`Fout bij het toevoegen van review: ${error.message || 'Onbekende fout'}`);
            console.error("Error adding review:", error);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return (
            <Container>
                <Spin size="large" tip="Productdetails laden..." />
            </Container>
        );
    }

    if (!product) {
        return (
            <Container>
                <p>Product niet gevonden.</p>
            </Container>
        );
    }

    return (
        <Container>
            {/* Hoofd Card voor productdetails */}
            <Card style={{ marginBottom: '20px' }}>
                <div style={{ textAlign: 'center', marginBottom: '20px' }}>
                    {/* Afbeeldingen Carousel (optioneel, anders gewoon Image) */}
                    {product.afbeeldingen && product.afbeeldingen.length > 0 ? (
                        // Gebruik Carousel voor meerdere afbeeldingen, anders gewoon een Image
                        product.afbeeldingen.length > 1 ? (
                            <Carousel autoplay dots={{ className: 'carousel-dots' }} style={{ maxWidth: '400px', margin: '0 auto' }}>
                                {product.afbeeldingen.map((imgUrl, index) => (
                                    <div key={index}>
                                        <Image
                                            src={imgUrl}
                                            alt={`${product.naam} afbeelding ${index + 1}`}
                                            style={{ maxHeight: '300px', objectFit: 'contain', width: '100%' }}
                                        />
                                    </div>
                                ))}
                            </Carousel>
                        ) : (
                            <Image
                                src={product.afbeeldingen[0]}
                                alt={product.naam || 'Product afbeelding'}
                                style={{ maxHeight: '300px', objectFit: 'contain', width: '100%', maxWidth: '400px' }}
                            />
                        )
                    ) : (
                        <div style={{ height: '200px', display: 'flex', alignItems: 'center', justifyContent: 'center', backgroundColor: '#f0f0f0', borderRadius: '8px' }}>
                            <Text type="secondary">Geen afbeeldingen beschikbaar</Text>
                        </div>
                    )}
                </div>

                <Title level={2} style={{ textAlign: 'center' }}>{product.naam || 'Onbekend Product'}</Title>
                <Paragraph style={{ textAlign: 'center' }}>EAN: {product.ean}</Paragraph>
                <Paragraph style={{ textAlign: 'center' }}>Merk: {product.merk || 'N/A'}</Paragraph>

                <Divider />

                <Title level={3}>Omschrijving</Title>
                <Paragraph>{product.omschrijving || 'Geen beschrijving beschikbaar.'}</Paragraph>

                {product.gewicht > 0 && (
                    <Paragraph><strong>Gewicht:</strong> {product.gewicht} kg</Paragraph>
                )}

                {/* Product Type */}
                {product.type && (product.type.naam || product.type.omschrijving) && (
                    <>
                        <Title level={3}>Product Type</Title>
                        <Paragraph><strong>Naam:</strong> {product.type.naam || 'N/A'}</Paragraph>
                        <Paragraph><strong>Omschrijving:</strong> {product.type.omschrijving || 'N/A'}</Paragraph>
                    </>
                )}

                {/* Categorieën */}
                {product.categorieen && product.categorieen.length > 0 && (
                    <>
                        <Title level={3}>Categorieën</Title>
                        <Paragraph>
                            {product.categorieen.map((cat) => (
                                <Tag key={cat.ID}>{cat.naam}</Tag>
                            ))}
                        </Paragraph>
                    </>
                )}

                {/* Specificaties */}
                {product.specificaties && product.specificaties.length > 0 && (
                    <>
                        <Title level={3}>Specificaties</Title>
                        <List
                            dataSource={product.specificaties}
                            renderItem={(spec) => (
                                <List.Item>
                                    <Text strong>{spec.naam}:</Text> {spec.waarde}
                                </List.Item>
                            )}
                        />
                    </>
                )}

                {/* Product Aanbod */}
                {product.productAanbod && product.productAanbod.length > 0 && (
                    <>
                        <Title level={3}>Aanbod</Title>
                        <Paragraph>
                            <strong>Vanaf Prijs:</strong> &euro;{getLowestPrice(product.productAanbod)}
                        </Paragraph>
                        <List
                            dataSource={product.productAanbod}
                            renderItem={(offer) => (
                                <List.Item>
                                    <Text>Prijs: &euro;{offer.prijs.toFixed(2)}</Text>
                                    <Text>Voorraad: {offer.voorraad} stuks</Text>
                                    <Text>Leverancier: {offer.supplier ? offer.supplier.name : 'N/A'}</Text>
                                </List.Item>
                            )}
                        />
                    </>
                )}

                {/* Tags */}
                {product.tags && product.tags.length > 0 && (
                    <>
                        <Title level={3}>Tags</Title>
                        <Paragraph>
                            {product.tags.map((tag) => (
                                <Tag key={tag.id}>{tag.naam}</Tag>
                            ))}
                        </Paragraph>
                    </>
                )}
            </Card>

            <Divider />

            {/* Reviews sectie */}
            <Card title="Reviews" style={{ marginTop: '20px' }}>
                {product.reviews && product.reviews.length > 0 ? (
                    <List
                        itemLayout="horizontal"
                        dataSource={product.reviews} // Gebruik product.reviews direct
                        renderItem={(review) => (
                            <List.Item>
                                <List.Item.Meta
                                    title={
                                        <>
                                            <Rate disabled defaultValue={review.score} />
                                            <span style={{ marginLeft: 8 }}>{review.titel || 'Geen titel'}</span>
                                        </>
                                    }
                                    description={
                                        <>
                                            <Paragraph>{review.inhoud || 'Geen inhoud beschikbaar.'}</Paragraph>
                                            <Text type="secondary">Door: {review.naam || 'Anoniem'}</Text>
                                        </>
                                    }
                                />
                            </List.Item>
                        )}
                    />
                ) : (
                    <Paragraph>Nog geen reviews voor dit product. Wees de eerste!</Paragraph>
                )}
            </Card>

            {/* Review formulier */}
            <Card title="Plaats een Review" style={{ marginTop: '20px' }}>
                <Form
                    form={form}
                    layout="vertical"
                    onFinish={handleAddReview}
                >
                    <Form.Item
                        name="name" // Naam van de reviewer
                        label="Naam"
                        rules={[{ required: true, message: 'Vul uw naam in!' }]}
                    >
                        <Input />
                    </Form.Item>
                    <Form.Item
                        name="score" // Beoordeling
                        label="Beoordeling"
                        rules={[{ required: true, message: 'Geef een beoordeling!' }]}
                    >
                        <Rate />
                    </Form.Item>
                    <Form.Item
                        name="title" // Titel van de review
                        label="Titel"
                        rules={[{ required: true, message: 'Voer een titel in!' }]}
                    >
                        <Input />
                    </Form.Item>
                    <Form.Item
                        name="content" // Inhoud van de review
                        label="Inhoud"
                        rules={[{ required: true, message: 'Voer een opmerking in!' }]}
                    >
                        <TextArea rows={4} />
                    </Form.Item>
                    <Form.Item>
                        <Button type="primary" htmlType="submit" loading={loading}>
                            Review Toevoegen
                        </Button>
                    </Form.Item>
                </Form>
            </Card>
        </Container>
    );
};

export default ProductDetailPage;