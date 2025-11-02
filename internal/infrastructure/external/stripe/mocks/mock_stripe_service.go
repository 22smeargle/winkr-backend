
package mocks

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v76"

	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/pkg/config"
)

// MockStripeService is a mock implementation of StripeService for testing
type MockStripeService struct {
	// Mock data
	customers       map[string]*stripe.Customer
	paymentIntents  map[string]*stripe.PaymentIntent
	subscriptions   map[string]*stripe.Subscription
	paymentMethods  map[string]*stripe.PaymentMethod
	prices          map[string]*stripe.Price
	invoices         map[string]*stripe.Invoice
	refunds          map[string]*stripe.Refund
	webhookEvents    map[string]interface{}
	
	// Error simulation
	simulateError bool
	errorMessage    string
}

// NewMockStripeService creates a new mock Stripe service
func NewMockStripeService() *MockStripeService {
	return &MockStripeService{
		customers:      make(map[string]*stripe.Customer),
		paymentIntents: make(map[string]*stripe.PaymentIntent),
		subscriptions:   make(map[string]*stripe.Subscription),
		paymentMethods:  make(map[string]*stripe.PaymentMethod),
		prices:         make(map[string]*stripe.Price),
		invoices:        make(map[string]*stripe.Invoice),
		refunds:         make(map[string]*stripe.Refund),
		webhookEvents:   make(map[string]interface{}),
	}
}

// SetSimulateError configures the mock to simulate errors
func (m *MockStripeService) SetSimulateError(simulate bool, message string) {
	m.simulateError = simulate
	m.errorMessage = message
}

// SetPlansResponse sets mock response for plans
func (m *MockStripeService) SetPlansResponse(prices []*stripe.Price, err error) {
	for _, price := range prices {
		m.prices[string(price.ID)] = price
	}
}

// SetCreateCustomerResponse sets mock response for customer creation
func (m *MockStripeService) SetCreateCustomerResponse(customer *stripe.Customer, err error) {
	if customer != nil {
		m.customers[customer.ID] = customer
	}
}

// SetGetCustomerResponse sets mock response for customer retrieval
func (m *MockStripeService) SetGetCustomerResponse(customer *stripe.Customer, err error) {
	if customer != nil {
		m.customers[customer.ID] = customer
	}
}

// SetUpdateCustomerResponse sets mock response for customer update
func (m *MockStripeService) SetUpdateCustomerResponse(customer *stripe.Customer, err error) {
	if customer != nil {
		m.customers[customer.ID] = customer
	}
}

// SetDeleteCustomerResponse sets mock response for customer deletion
func (m *MockStripeService) SetDeleteCustomerResponse(err error) {
	// Mock implementation
}

// SetCreatePaymentIntentResponse sets mock response for payment intent creation
func (m *MockStripeService) SetCreatePaymentIntentResponse(paymentIntent *stripe.PaymentIntent, err error) {
	if paymentIntent != nil {
		m.paymentIntents[paymentIntent.ID] = paymentIntent
	}
}

// SetConfirmPaymentIntentResponse sets mock response for payment intent confirmation
func (m *MockStripeService) SetConfirmPaymentIntentResponse(paymentIntent *stripe.PaymentIntent, err error) {
	if paymentIntent != nil {
		m.paymentIntents[paymentIntent.ID] = paymentIntent
	}
}

// SetCreateSubscriptionResponse sets mock response for subscription creation
func (m *MockStripeService) SetCreateSubscriptionResponse(subscription *stripe.Subscription, err error) {
	if subscription != nil {
		m.subscriptions[subscription.ID] = subscription
	}
}

// SetGetSubscriptionResponse sets mock response for subscription retrieval
func (m *MockStripeService) SetGetSubscriptionResponse(subscription *stripe.Subscription, err error) {
	if subscription != nil {
		m.subscriptions[subscription.ID] = subscription
	}
}

// SetUpdateSubscriptionResponse sets mock response for subscription update
func (m *MockStripeService) SetUpdateSubscriptionResponse(subscription *stripe.Subscription, err error) {
	if subscription != nil {
		m.subscriptions[subscription.ID] = subscription
	}
}

// SetCancelSubscriptionResponse sets mock response for subscription cancellation
func (m *MockStripeService) SetCancelSubscriptionResponse(subscription *stripe.Subscription, err error) {
	if subscription != nil {
		m.subscriptions[subscription.ID] = subscription
	}
}

// SetCreatePaymentMethodResponse sets mock response for payment method creation
func (m *MockStripeService) SetCreatePaymentMethodResponse(paymentMethod *stripe.PaymentMethod, err error) {
	if paymentMethod != nil {
		m.paymentMethods[paymentMethod.ID] = paymentMethod
	}
}

// SetGetPaymentMethodResponse sets mock response for payment method retrieval
func (m *MockStripeService) SetGetPaymentMethodResponse(paymentMethod *stripe.PaymentMethod, err error) {
	if paymentMethod != nil {
		m.paymentMethods[paymentMethod.ID] = paymentMethod
	}
}

// SetListPaymentMethodsResponse sets mock response for payment methods listing
func (m *MockStripeService) SetListPaymentMethodsResponse(paymentMethods *stripe.PaymentMethodList, err error) {
	if paymentMethods != nil {
		for _, pm := range paymentMethods.Data {
			m.paymentMethods[pm.ID] = pm
		}
	}
}

// SetDetachPaymentMethodResponse sets mock response for payment method detachment
func (m *MockStripeService) SetDetachPaymentMethodResponse(err error) {
	// Mock implementation
}

// SetCreateRefundResponse sets mock response for refund creation
func (m *MockStripeService) SetCreateRefundResponse(refund *stripe.Refund, err error) {
	if refund != nil {
		m.refunds[refund.ID] = refund
	}
}

// SetGetInvoiceResponse sets mock response for invoice retrieval
func (m *MockStripeService) SetGetInvoiceResponse(invoice *stripe.Invoice, err error) {
	if invoice != nil {
		m.invoices[invoice.ID] = invoice
	}
}

// SetListInvoicesResponse sets mock response for invoices listing
func (m *MockStripeService) SetListInvoicesResponse(invoices *stripe.InvoiceList, err error) {
	if invoices != nil {
		for _, invoice := range invoices.Data {
			m.invoices[invoice.ID] = invoice
		}
	}
}

// SetProcessEventResponse sets mock response for webhook event processing
func (m *MockStripeService) SetProcessEventResponse(err error) {
	// Mock implementation
}

// Mock StripeService interface methods

func (m *MockStripeService) CreateCustomer(ctx context.Context, params *stripe.CustomerParams) (*stripe.Customer, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	customer := &stripe.Customer{
		ID:    "cus_mock_" + params.Email,
		Email: params.Email,
		Name:  params.Name,
	}
	
	m.customers[customer.ID] = customer
	return customer, nil
}

func (m *MockStripeService) GetCustomer(ctx context.Context, customerID string) (*stripe.Customer, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if customer, exists := m.customers[customerID]; exists {
		return customer, nil
	}
	
	return nil, &stripe.Error{Msg: "Customer not found"}
}

func (m *MockStripeService) UpdateCustomer(ctx context.Context, customerID string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if customer, exists := m.customers[customerID]; exists {
		if params.Email != "" {
			customer.Email = params.Email
		}
		if params.Name != "" {
			customer.Name = params.Name
		}
		
		m.customers[customerID] = customer
		return customer, nil
	}
	
	return nil, &stripe.Error{Msg: "Customer not found"}
}

func (m *MockStripeService) DeleteCustomer(ctx context.Context, customerID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	delete(m.customers, customerID)
	return nil
}

func (m *MockStripeService) CreatePaymentIntent(ctx context.Context, params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	paymentIntent := &stripe.PaymentIntent{
		ID:     "pi_mock_" + params.Customer.ID,
		Amount: params.Amount,
		Currency: stripe.String(string(params.Currency)),
		Status:  "requires_payment_method",
		Customer: params.Customer,
	}
	
	m.paymentIntents[paymentIntent.ID] = paymentIntent
	return paymentIntent, nil
}

func (m *MockStripeService) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string, params *stripe.PaymentIntentConfirmParams) (*stripe.PaymentIntent, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if paymentIntent, exists := m.paymentIntents[paymentIntentID]; exists {
		paymentIntent.Status = "succeeded"
		m.paymentIntents[paymentIntentID] = paymentIntent
		return paymentIntent, nil
	}
	
	return nil, &stripe.Error{Msg: "Payment intent not found"}
}

func (m *MockStripeService) CreateSubscription(ctx context.Context, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	subscription := &stripe.Subscription{
		ID:       "sub_mock_" + params.Customer.ID,
		Customer: params.Customer,
		Items:    params.Items,
		Status:    "active",
	}
	
	m.subscriptions[subscription.ID] = subscription
	return subscription, nil
}

func (m *MockStripeService) GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		return subscription, nil
	}
	
	return nil, &stripe.Error{Msg: "Subscription not found"}
}

func (m *MockStripeService) UpdateSubscription(ctx context.Context, subscriptionID string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		if params.Items != nil {
			subscription.Items = params.Items
		}
		
		m.subscriptions[subscriptionID] = subscription
		return subscription, nil
	}
	
	return nil, &stripe.Error{Msg: "Subscription not found"}
}

func (m *MockStripeService) CancelSubscription(ctx context.Context, subscriptionID string, params *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if subscription, exists := m.subscriptions[subscriptionID]; exists {
		subscription.Status = "canceled"
		if params.CancelAtPeriodEnd {
			subscription.CancelAtPeriodEnd = true
		}
		
		m.subscriptions[subscriptionID] = subscription
		return subscription, nil
	}
	
	return nil, &stripe.Error{Msg: "Subscription not found"}
}

func (m *MockStripeService) CreatePaymentMethod(ctx context.Context, params *stripe.PaymentMethodParams) (*stripe.PaymentMethod, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	paymentMethod := &stripe.PaymentMethod{
		ID:   "pm_mock_" + params.Customer.ID,
		Type:  "card",
		Card: &stripe.PaymentMethodCard{
			Brand:  "visa",
			Last4:  "4242",
			ExpMonth: 12,
			ExpYear: 2025,
		},
		Customer: params.Customer,
	}
	
	m.paymentMethods[paymentMethod.ID] = paymentMethod
	return paymentMethod, nil
}

func (m *MockStripeService) GetPaymentMethod(ctx context.Context, paymentMethodID string) (*stripe.PaymentMethod, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if paymentMethod, exists := m.paymentMethods[paymentMethodID]; exists {
		return paymentMethod, nil
	}
	
	return nil, &stripe.Error{Msg: "Payment method not found"}
}

func (m *MockStripeService) ListPaymentMethods(ctx context.Context, params *stripe.PaymentMethodListParams) (*stripe.PaymentMethodList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	var paymentMethods []*stripe.PaymentMethod
	for _, pm := range m.paymentMethods {
		if params.Customer != nil && pm.Customer.ID == params.Customer.ID {
			paymentMethods = append(paymentMethods, pm)
		}
	}
	
	return &stripe.PaymentMethodList{
		Data: paymentMethods,
	}, nil
}

func (m *MockStripeService) DetachPaymentMethod(ctx context.Context, paymentMethodID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	delete(m.paymentMethods, paymentMethodID)
	return nil
}

func (m *MockStripeService) CreateRefund(ctx context.Context, params *stripe.RefundParams) (*stripe.Refund, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	refund := &stripe.Refund{
		ID:     "re_mock_" + params.PaymentIntent.ID,
		Amount: params.Amount,
		Status: "succeeded",
	}
	
	m.refunds[refund.ID] = refund
	return refund, nil
}

func (m *MockStripeService) GetInvoice(ctx context.Context, invoiceID string) (*stripe.Invoice, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if invoice, exists := m.invoices[invoiceID]; exists {
		return invoice, nil
	}
	
	return nil, &stripe.Error{Msg: "Invoice not found"}
}

func (m *MockStripeService) ListInvoices(ctx context.Context, params *stripe.InvoiceListParams) (*stripe.InvoiceList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	var invoices []*stripe.Invoice
	for _, invoice := range m.invoices {
		if params.Customer != nil && invoice.Customer.ID == params.Customer.ID {
			invoices = append(invoices, invoice)
		}
	}
	
	return &stripe.InvoiceList{
		Data: invoices,
	}, nil
}

func (m *MockStripeService) GetPrices(ctx context.Context, params *stripe.PriceListParams) (*stripe.PriceList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	var prices []*stripe.Price
	for _, price := range m.prices {
		prices = append(prices, price)
	}
	
	return &stripe.PriceList{
		Data: prices,
	}, nil
}

func (m *MockStripeService) VerifyWebhookSignature(payload []byte, signature string, secret string) error {
	// Mock implementation - always return success for testing
	return nil
}

func (m *MockStripeService) ProcessWebhookEvent(ctx context.Context, event stripe.Event) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Store the event for testing
	m.webhookEvents[event.ID] = event
	return nil
}

func (m *MockStripeService) GetCustomerSubscriptions(ctx context.Context, customerID string, params *stripe.SubscriptionListParams) (*stripe.SubscriptionList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	var subscriptions []*stripe.Subscription
	for _, subscription := range m.subscriptions {
		if subscription.Customer.ID == customerID {
			subscriptions = append(subscriptions, subscription)
		}
	}
	
	return &stripe.SubscriptionList{
		Data: subscriptions,
	}, nil
}

func (m *MockStripeService) GetUpcomingInvoices(ctx context.Context, params *stripe.InvoiceListParams) (*stripe.InvoiceList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	var invoices []*stripe.Invoice
	for _, invoice := range m.invoices {
		if params.Customer != nil && invoice.Customer.ID == params.Customer.ID {
			invoices = append(invoices, invoice)
		}
	}
	
	return &stripe.InvoiceList{
		Data: invoices,
	}, nil
}

func (m *MockStripeService) CreateInvoice(ctx context.Context, params *stripe.InvoiceParams) (*stripe.Invoice, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	invoice := &stripe.Invoice{
		ID:       "in_mock_" + params.Customer.ID,
		Customer: params.Customer,
		Status:   "draft",
	}
	
	m.invoices[invoice.ID] = invoice
	return invoice, nil
}

func (m *MockStripeService) FinalizeInvoice(ctx context.Context, invoiceID string, params *stripe.InvoiceFinalizeParams) (*stripe.Invoice, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if invoice, exists := m.invoices[invoiceID]; exists {
		invoice.Status = "open"
		m.invoices[invoiceID] = invoice
		return invoice, nil
	}
	
	return nil, &stripe.Error{Msg: "Invoice not found"}
}

func (m *MockStripeService) PayInvoice(ctx context.Context, invoiceID string, params *stripe.InvoicePayParams) (*stripe.Invoice, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if invoice, exists := m.invoices[invoiceID]; exists {
		invoice.Status = "paid"
		m.invoices[invoiceID] = invoice
		return invoice, nil
	}
	
	return nil, &stripe.Error{Msg: "Invoice not found"}
}

func (m *MockStripeService) SendInvoice(ctx context.Context, invoiceID string, params *stripe.InvoiceSendParams) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	return nil
}

func (m *MockStripeService) VoidInvoice(ctx context.Context, invoiceID string, params *stripe.InvoiceVoidParams) (*stripe.Invoice, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if invoice, exists := m.invoices[invoiceID]; exists {
		invoice.Status = "void"
		m.invoices[invoiceID] = invoice
		return invoice, nil
	}
	
	return nil, &stripe.Error{Msg: "Invoice not found"}
}

func (m *MockStripeService) UncollectInvoice(ctx context.Context, invoiceID string, params *stripe.InvoiceUncollectibleParams) (*stripe.Invoice, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	if invoice, exists := m.invoices[invoiceID]; exists {
		invoice.Status = "uncollectible"
		m.invoices[invoiceID] = invoice
		return invoice, nil
	}
	
	return nil, &stripe.Error{Msg: "Invoice not found"}
}

func (m *MockStripeService) CreateProduct(ctx context.Context, params *stripe.ProductParams) (*stripe.Product, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	product := &stripe.Product{
		ID:   "prod_mock",
		Name:  params.Name,
	}
	
	return product, nil
}

func (m *MockStripeService) UpdateProduct(ctx context.Context, productID string, params *stripe.ProductParams) (*stripe.Product, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	product := &stripe.Product{
		ID:   productID,
		Name:  params.Name,
	}
	
	return product, nil
}

func (m *MockStripeService) DeleteProduct(ctx context.Context, productID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) CreatePrice(ctx context.Context, params *stripe.PriceParams) (*stripe.Price, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	price := &stripe.Price{
		ID:       "price_mock_" + params.Product.ID,
		Product:  params.Product,
		UnitAmount: params.UnitAmount,
		Currency:  params.Currency,
	}
	
	return price, nil
}

func (m *MockStripeService) UpdatePrice(ctx context.Context, priceID string, params *stripe.PriceParams) (*stripe.Price, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	price := &stripe.Price{
		ID:       priceID,
		Product:  params.Product,
		UnitAmount: params.UnitAmount,
		Currency:  params.Currency,
	}
	
	return price, nil
}

func (m *MockStripeService) DeletePrice(ctx context.Context, priceID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) GetProduct(ctx context.Context, productID string) (*stripe.Product, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	product := &stripe.Product{
		ID:   productID,
		Name:  "Mock Product",
	}
	
	return product, nil
}

func (m *MockStripeService) ListProducts(ctx context.Context, params *stripe.ProductListParams) (*stripe.ProductList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	products := []*stripe.Product{
		{
			ID:   "prod_free",
			Name:  "Free Plan",
		},
		{
			ID:   "prod_premium",
			Name:  "Premium Plan",
		},
		{
			ID:   "prod_platinum",
			Name:  "Platinum Plan",
		},
	}
	
	return &stripe.ProductList{
		Data: products,
	}, nil
}

func (m *MockStripeService) CreateCheckoutSession(ctx context.Context, params *stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.CheckoutSession{
		ID: "cs_mock_" + params.Customer.ID,
		URL: "https://checkout.stripe.com/pay/cs_mock_" + params.Customer.ID,
	}
	
	return session, nil
}

func (m *MockStripeService) GetCheckoutSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.CheckoutSession{
		ID:  sessionID,
		URL: "https://checkout.stripe.com/pay/" + sessionID,
	}
	
	return session, nil
}

func (m *MockStripeService) CreateSetupIntent(ctx context.Context, params *stripe.SetupIntentParams) (*stripe.SetupIntent, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	intent := &stripe.SetupIntent{
		ID:     "seti_mock_" + params.Customer.ID,
		Status: "requires_payment_method",
	}
	
	return intent, nil
}

func (m *MockStripeService) ConfirmSetupIntent(ctx context.Context, setupIntentID string, params *stripe.SetupIntentConfirmParams) (*stripe.SetupIntent, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	intent := &stripe.SetupIntent{
		ID:     setupIntentID,
		Status: "succeeded",
	}
	
	return intent, nil
}

func (m *MockStripeService) GetSetupIntent(ctx context.Context, setupIntentID string) (*stripe.SetupIntent, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	intent := &stripe.SetupIntent{
		ID:     setupIntentID,
		Status: "succeeded",
	}
	
	return intent, nil
}

func (m *MockStripeService) CreateTaxRate(ctx context.Context, params *stripe.TaxRateParams) (*stripe.TaxRate, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	taxRate := &stripe.TaxRate{
		ID:       "txr_mock",
		DisplayName: "Mock Tax Rate",
		Percentage: 10.0,
		Inclusive: false,
	}
	
	return taxRate, nil
}

func (m *MockStripeService) GetTaxRate(ctx context.Context, taxRateID string) (*stripe.TaxRate, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	taxRate := &stripe.TaxRate{
		ID:       taxRateID,
		DisplayName: "Mock Tax Rate",
		Percentage: 10.0,
		Inclusive: false,
	}
	
	return taxRate, nil
}

func (m *MockStripeService) ListTaxRates(ctx context.Context, params *stripe.TaxRateListParams) (*stripe.TaxRateList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	taxRates := []*stripe.TaxRate{
		{
			ID:       "txr_mock_1",
			DisplayName: "Mock Tax Rate 1",
			Percentage: 10.0,
			Inclusive: false,
		},
		{
			ID:       "txr_mock_2",
			DisplayName: "Mock Tax Rate 2",
			Percentage: 20.0,
			Inclusive: false,
		},
	}
	
	return &stripe.TaxRateList{
		Data: taxRates,
	}, nil
}

func (m *MockStripeService) UpdateTaxRate(ctx context.Context, taxRateID string, params *stripe.TaxRateParams) (*stripe.TaxRate, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	taxRate := &stripe.TaxRate{
		ID:       taxRateID,
		DisplayName: params.DisplayName,
		Percentage: params.Percentage,
		Inclusive: params.Inclusive,
	}
	
	return taxRate, nil
}

func (m *MockStripeService) DeleteTaxRate(ctx context.Context, taxRateID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) CreateCoupon(ctx context.Context, params *stripe.CouponParams) (*stripe.Coupon, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	coupon := &stripe.Coupon{
		ID:         "coupon_mock",
		PercentOff: 10.0,
		Duration:   "repeating",
	}
	
	return coupon, nil
}

func (m *MockStripeService) GetCoupon(ctx context.Context, couponID string) (*stripe.Coupon, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	coupon := &stripe.Coupon{
		ID:         couponID,
		PercentOff: 10.0,
		Duration:   "repeating",
	}
	
	return coupon, nil
}

func (m *MockStripeService) UpdateCoupon(ctx context.Context, couponID string, params *stripe.CouponParams) (*stripe.Coupon, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	coupon := &stripe.Coupon{
		ID:         couponID,
		PercentOff: params.PercentOff,
		Duration:   params.Duration,
	}
	
	return coupon, nil
}

func (m *MockStripeService) DeleteCoupon(ctx context.Context, couponID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListCoupons(ctx context.Context, params *stripe.CouponListParams) (*stripe.CouponList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	coupons := []*stripe.Coupon{
		{
			ID:         "coupon_mock_1",
			PercentOff: 10.0,
			Duration:   "repeating",
		},
		{
			ID:         "coupon_mock_2",
			PercentOff: 20.0,
			Duration:   "repeating",
		},
	}
	
	return &stripe.CouponList{
		Data: coupons,
	}, nil
}

func (m *MockStripeService) CreatePromotionCode(ctx context.Context, params *stripe.PromotionCodeParams) (*stripe.PromotionCode, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	promotionCode := &stripe.PromotionCode{
		ID:       "promo_mock",
		Code:     params.Code,
		Active:   true,
	}
	
	return promotionCode, nil
}

func (m *MockStripeService) GetPromotionCode(ctx context.Context, promotionCodeID string) (*stripe.PromotionCode, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	promotionCode := &stripe.PromotionCode{
		ID:   promotionCodeID,
		Code: "PROMO10",
		Active: true,
	}
	
	return promotionCode, nil
}

func (m *MockStripeService) UpdatePromotionCode(ctx context.Context, promotionCodeID string, params *stripe.PromotionCodeParams) (*stripe.PromotionCode, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	promotionCode := &stripe.PromotionCode{
		ID:   promotionCodeID,
		Code: params.Code,
		Active: params.Active,
	}
	
	return promotionCode, nil
}

func (m *MockStripeService) DeletePromotionCode(ctx context.Context, promotionCodeID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListPromotionCodes(ctx context.Context, params *stripe.PromotionCodeListParams) (*stripe.PromotionCodeList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	promotionCodes := []*stripe.PromotionCode{
		{
			ID:     "promo_mock_1",
			Code:   "PROMO10",
			Active: true,
		},
		{
			ID:     "promo_mock_2",
			Code:   "PROMO20",
			Active: true,
		},
	}
	
	return &stripe.PromotionCodeList{
		Data: promotionCodes,
	}, nil
}

func (m *MockStripeService) CreateBalanceTransaction(ctx context.Context, params *stripe.BalanceTransactionParams) (*stripe.BalanceTransaction, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	transaction := &stripe.BalanceTransaction{
		ID:     "txn_mock",
		Amount: params.Amount,
		Type:   "payment",
	}
	
	return transaction, nil
}

func (m *MockStripeService) GetBalance(ctx context.Context, params *stripe.BalanceParams) (*stripe.Balance, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	balance := &stripe.Balance{
		Available: 10000, // $100.00
		Pending:   0,
	}
	
	return balance, nil
}

func (m *MockStripeService) CreateBalanceTransaction(ctx context.Context, params *stripe.BalanceTransactionParams) (*stripe.BalanceTransaction, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	transaction := &stripe.BalanceTransaction{
		ID:     "txn_mock",
		Amount: params.Amount,
		Type:   "payment",
	}
	
	return transaction, nil
}

func (m *MockStripeService) GetBalanceTransaction(ctx context.Context, transactionID string) (*stripe.BalanceTransaction, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	transaction := &stripe.BalanceTransaction{
		ID:     transactionID,
		Amount: 1000, // $10.00
		Type:   "payment",
	}
	
	return transaction, nil
}

func (m *MockStripeService) ListBalanceTransactions(ctx context.Context, params *stripe.BalanceTransactionListParams) (*stripe.BalanceTransactionList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	transactions := []*stripe.BalanceTransaction{
		{
			ID:     "txn_mock_1",
			Amount: 1000, // $10.00
			Type:   "payment",
		},
		{
			ID:     "txn_mock_2",
			Amount: 2000, // $20.00
			Type:   "payout",
		},
	}
	
	return &stripe.BalanceTransactionList{
		Data: transactions,
	}, nil
}

func (m *MockStripeService) CreateAccount(ctx context.Context, params *stripe.AccountParams) (*stripe.Account, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	account := &stripe.Account{
		ID:   "acct_mock",
		Type: "custom",
	}
	
	return account, nil
}

func (m *MockStripeService) GetAccount(ctx context.Context, accountID string) (*stripe.Account, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	account := &stripe.Account{
		ID:   accountID,
		Type: "custom",
	}
	
	return account, nil
}

func (m *MockStripeService) UpdateAccount(ctx context.Context, accountID string, params *stripe.AccountParams) (*stripe.Account, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	account := &stripe.Account{
		ID:   accountID,
		Type: "custom",
	}
	
	return account, nil
}

func (m *MockStripeService) DeleteAccount(ctx context.Context, accountID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListAccounts(ctx context.Context, params *stripe.AccountListParams) (*stripe.AccountList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	accounts := []*stripe.Account{
		{
			ID:   "acct_mock_1",
			Type: "custom",
		},
		{
			ID:   "acct_mock_2",
			Type: "express",
		},
	}
	
	return &stripe.AccountList{
		Data: accounts,
	}, nil
}

func (m *MockStripeService) CreateAccountLink(ctx context.Context, params *stripe.AccountLinkParams) (*stripe.AccountLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	accountLink := &stripe.AccountLink{
		ID:   "acct_link_mock",
		URL:  "https://connect.stripe.com/express/mock",
	}
	
	return accountLink, nil
}

func (m *MockStripeService) GetAccountLink(ctx context.Context, accountLinkID string) (*stripe.AccountLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	accountLink := &stripe.AccountLink{
		ID:  accountLinkID,
		URL: "https://connect.stripe.com/express/mock",
	}
	
	return accountLink, nil
}

func (m *MockStripeService) UpdateAccountLink(ctx context.Context, accountLinkID string, params *stripe.AccountLinkParams) (*stripe.AccountLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	accountLink := &stripe.AccountLink{
		ID:  accountLinkID,
		URL: "https://connect.stripe.com/express/mock",
	}
	
	return accountLink, nil
}

func (m *MockStripeService) DeleteAccountLink(ctx context.Context, accountLinkID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListAccountLinks(ctx context.Context, params *stripe.AccountLinkListParams) (*stripe.AccountLinkList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	accountLinks := []*stripe.AccountLink{
		{
			ID:  "acct_link_mock_1",
			URL: "https://connect.stripe.com/express/mock1",
		},
		{
			ID:  "acct_link_mock_2",
			URL: "https://connect.stripe.com/express/mock2",
		},
	}
	
	return &stripe.AccountLinkList{
		Data: accountLinks,
	}, nil
}

func (m *MockStripeService) CreateLoginLink(ctx context.Context, params *stripe.LoginLinkParams) (*stripe.LoginLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	loginLink := &stripe.LoginLink{
		ID:  "login_link_mock",
		URL: "https://connect.stripe.com/login/mock",
	}
	
	return loginLink, nil
}

func (m *MockStripeService) GetLoginLink(ctx context.Context, loginLinkID string) (*stripe.LoginLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	loginLink := &stripe.LoginLink{
		ID:  loginLinkID,
		URL: "https://connect.stripe.com/login/mock",
	}
	
	return loginLink, nil
}

func (m *MockStripeService) UpdateLoginLink(ctx context.Context, loginLinkID string, params *stripe.LoginLinkParams) (*stripe.LoginLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	loginLink := &stripe.LoginLink{
		ID:  loginLinkID,
		URL: "https://connect.stripe.com/login/mock",
	}
	
	return loginLink, nil
}

func (m *MockStripeService) DeleteLoginLink(ctx context.Context, loginLinkID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListLoginLinks(ctx context.Context, params *stripe.LoginLinkListParams) (*stripe.LoginLinkList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	loginLinks := []*stripe.LoginLink{
		{
			ID:  "login_link_mock_1",
			URL: "https://connect.stripe.com/login/mock1",
		},
		{
			ID:  "login_link_mock_2",
			URL: "https://connect.stripe.com/login/mock2",
		},
	}
	
	return &stripe.LoginLinkList{
		Data: loginLinks,
	}, nil
}

func (m *MockStripeService) CreateFile(ctx context.Context, params *stripe.FileParams) (*stripe.File, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	file := &stripe.File{
		ID:   "file_mock",
		Purpose: "identity_document",
	}
	
	return file, nil
}

func (m *MockStripeService) GetFile(ctx context.Context, fileID string) (*stripe.File, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	file := &stripe.File{
		ID:   fileID,
		Purpose: "identity_document",
	}
	
	return file, nil
}

func (m *MockStripeService) UpdateFile(ctx context.Context, fileID string, params *stripe.FileParams) (*stripe.File, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	file := &stripe.File{
		ID:   fileID,
		Purpose: "identity_document",
	}
	
	return file, nil
}

func (m *MockStripeService) DeleteFile(ctx context.Context, fileID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListFiles(ctx context.Context, params *stripe.FileListParams) (*stripe.FileList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	files := []*stripe.File{
		{
			ID:     "file_mock_1",
			Purpose: "identity_document",
		},
		{
			ID:     "file_mock_2",
			Purpose: "dispute_evidence",
		},
	}
	
	return &stripe.FileList{
		Data: files,
	}, nil
}

func (m *MockStripeService) CreateFileLink(ctx context.Context, params *stripe.FileLinkParams) (*stripe.FileLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	fileLink := &stripe.FileLink{
		ID:  "file_link_mock",
		URL: "https://files.stripe.com/link/mock",
	}
	
	return fileLink, nil
}

func (m *MockStripeService) GetFileLink(ctx context.Context, fileLinkID string) (*stripe.FileLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	fileLink := &stripe.FileLink{
		ID:  fileLinkID,
		URL: "https://files.stripe.com/link/mock",
	}
	
	return fileLink, nil
}

func (m *MockStripeService) UpdateFileLink(ctx context.Context, fileLinkID string, params *stripe.FileLinkParams) (*stripe.FileLink, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	fileLink := &stripe.FileLink{
		ID:  fileLinkID,
		URL: "https://files.stripe.com/link/mock",
	}
	
	return fileLink, nil
}

func (m *MockStripeService) DeleteFileLink(ctx context.Context, fileLinkID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListFileLinks(ctx context.Context, params *stripe.FileLinkListParams) (*stripe.FileLinkList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	fileLinks := []*stripe.FileLink{
		{
			ID:  "file_link_mock_1",
			URL: "https://files.stripe.com/link/mock1",
		},
		{
			ID:  "file_link_mock_2",
			URL: "https://files.stripe.com/link/mock2",
		},
	}
	
	return &stripe.FileLinkList{
		Data: fileLinks,
	}, nil
}

func (m *MockStripeService) CreateToken(ctx context.Context, params *stripe.TokenParams) (*stripe.Token, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	token := &stripe.Token{
		ID:    "tok_mock",
		Type:  "card",
		Card: &stripe.PaymentMethodCard{
			Brand:  "visa",
			Last4:  "4242",
			ExpMonth: 12,
			ExpYear: 2025,
		},
	}
	
	return token, nil
}

func (m *MockStripeService) GetToken(ctx context.Context, tokenID string) (*stripe.Token, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	token := &stripe.Token{
		ID:    tokenID,
		Type:  "card",
		Card: &stripe.PaymentMethodCard{
			Brand:  "visa",
			Last4:  "4242",
			ExpMonth: 12,
			ExpYear: 2025,
		},
	}
	
	return token, nil
}

func (m *MockStripeService) CreatePayout(ctx context.Context, params *stripe.PayoutParams) (*stripe.Payout, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	payout := &stripe.Payout{
		ID:     "po_mock",
		Amount:  params.Amount,
		Status: "in_transit",
	}
	
	return payout, nil
}

func (m *MockStripeService) GetPayout(ctx context.Context, payoutID string) (*stripe.Payout, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	payout := &stripe.Payout{
		ID:     payoutID,
		Amount:  10000, // $100.00
		Status: "in_transit",
	}
	
	return payout, nil
}

func (m *MockStripeService) UpdatePayout(ctx context.Context, payoutID string, params *stripe.PayoutParams) (*stripe.Payout, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	payout := &stripe.Payout{
		ID:     payoutID,
		Amount:  params.Amount,
		Status: "in_transit",
	}
	
	return payout, nil
}

func (m *MockStripeService) DeletePayout(ctx context.Context, payoutID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListPayouts(ctx context.Context, params *stripe.PayoutListParams) (*stripe.PayoutList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	payouts := []*stripe.Payout{
		{
			ID:     "po_mock_1",
			Amount:  10000, // $100.00
			Status: "in_transit",
		},
		{
			ID:     "po_mock_2",
			Amount:  20000, // $200.00
			Status: "paid",
		},
	}
	
	return &stripe.PayoutList{
		Data: payouts,
	}, nil
}

func (m *MockStripeService) CreateTransfer(ctx context.Context, params *stripe.TransferParams) (*stripe.Transfer, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	transfer := &stripe.Transfer{
		ID:     "tr_mock",
		Amount:  params.Amount,
		Status: "in_transit",
	}
	
	return transfer, nil
}

func (m *MockStripeService) GetTransfer(ctx context.Context, transferID string) (*stripe.Transfer, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	transfer := &stripe.Transfer{
		ID:     transferID,
		Amount:  10000, // $100.00
		Status: "in_transit",
	}
	
	return transfer, nil
}

func (m *MockStripeService) UpdateTransfer(ctx context.Context, transferID string, params *stripe.TransferParams) (*stripe.Transfer, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	transfer := &stripe.Transfer{
		ID:     transferID,
		Amount:  params.Amount,
		Status: "in_transit",
	}
	
	return transfer, nil
}

func (m *MockStripeService) DeleteTransfer(ctx context.Context, transferID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListTransfers(ctx context.Context, params *stripe.TransferListParams) (*stripe.TransferList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	transfers := []*stripe.Transfer{
		{
			ID:     "tr_mock_1",
			Amount:  10000, // $100.00
			Status: "in_transit",
		},
		{
			ID:     "tr_mock_2",
			Amount:  20000, // $200.00
			Status: "paid",
		},
	}
	
	return &stripe.TransferList{
		Data: transfers,
	}, nil
}

func (m *MockStripeService) CreateTransferReversal(ctx context.Context, params *stripe.TransferReversalParams) (*stripe.TransferReversal, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	reversal := &stripe.TransferReversal{
		ID:     "tr_rev_mock",
		Amount:  params.Amount,
		Status: "succeeded",
	}
	
	return reversal, nil
}

func (m *MockStripeService) GetTransferReversal(ctx context.Context, reversalID string) (*stripe.TransferReversal, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	reversal := &stripe.TransferReversal{
		ID:     reversalID,
		Amount:  10000, // $100.00
		Status: "succeeded",
	}
	
	return reversal, nil
}

func (m *MockStripeService) UpdateTransferReversal(ctx context.Context, reversalID string, params *stripe.TransferReversalParams) (*stripe.TransferReversal, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	reversal := &stripe.TransferReversal{
		ID:     reversalID,
		Amount:  params.Amount,
		Status: "succeeded",
	}
	
	return reversal, nil
}

func (m *MockStripeService) DeleteTransferReversal(ctx context.Context, reversalID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListTransferReversals(ctx context.Context, params *stripe.TransferReversalListParams) (*stripe.TransferReversalList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	reversals := []*stripe.TransferReversal{
		{
			ID:     "tr_rev_mock_1",
			Amount:  10000, // $100.00
			Status: "succeeded",
		},
		{
			ID:     "tr_rev_mock_2",
			Amount:  20000, // $200.00
			Status: "failed",
		},
	}
	
	return &stripe.TransferReversalList{
		Data: reversals,
	}, nil
}

func (m *MockStripeService) CreateApplication(ctx context.Context, params *stripe.ApplicationParams) (*stripe.Application, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	application := &stripe.Application{
		ID:   "ca_mock",
		Name: "Mock Application",
	}
	
	return application, nil
}

func (m *MockStripeService) GetApplication(ctx context.Context, applicationID string) (*stripe.Application, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	application := &stripe.Application{
		ID:   applicationID,
		Name: "Mock Application",
	}
	
	return application, nil
}

func (m *MockStripeService) UpdateApplication(ctx context.Context, applicationID string, params *stripe.ApplicationParams) (*stripe.Application, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	application := &stripe.Application{
		ID:   applicationID,
		Name: params.Name,
	}
	
	return application, nil
}

func (m *MockStripeService) DeleteApplication(ctx context.Context, applicationID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListApplications(ctx context.Context, params *stripe.ApplicationListParams) (*stripe.ApplicationList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	applications := []*stripe.Application{
		{
			ID:   "ca_mock_1",
			Name: "Mock Application 1",
		},
		{
			ID:   "ca_mock_2",
			Name: "Mock Application 2",
		},
	}
	
	return &stripe.ApplicationList{
		Data: applications,
	}, nil
}

func (m *MockStripeService) CreateApplicationFee(ctx context.Context, params *stripe.ApplicationFeeParams) (*stripe.ApplicationFee, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	fee := &stripe.ApplicationFee{
		ID:     "fee_mock",
		Amount:  params.Amount,
		Type:   "application",
	}
	
	return fee, nil
}

func (m *MockStripeService) GetApplicationFee(ctx context.Context, feeID string) (*stripe.ApplicationFee, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	fee := &stripe.ApplicationFee{
		ID:     feeID,
		Amount:  100, // $1.00
		Type:   "application",
	}
	
	return fee, nil
}

func (m *MockStripeService) UpdateApplicationFee(ctx context.Context, feeID string, params *stripe.ApplicationFeeParams) (*stripe.ApplicationFee, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	fee := &stripe.ApplicationFee{
		ID:     feeID,
		Amount:  params.Amount,
		Type:   params.Type,
	}
	
	return fee, nil
}

func (m *MockStripeService) DeleteApplicationFee(ctx context.Context, feeID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListApplicationFees(ctx context.Context, params *stripe.ApplicationFeeListParams) (*stripe.ApplicationFeeList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	fees := []*stripe.ApplicationFee{
		{
			ID:     "fee_mock_1",
			Amount:  100, // $1.00
			Type:   "application",
		},
		{
			ID:     "fee_mock_2",
			Amount:  200, // $2.00
			Type:   "application",
		},
	}
	
	return &stripe.ApplicationFeeList{
		Data: fees,
	}, nil
}

func (m *MockStripeService) CreateCapability(ctx context.Context, params *stripe.CapabilityParams) (*stripe.Capability, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	capability := &stripe.Capability{
		ID:   "cap_mock",
		Key:  "card_payments",
	}
	
	return capability, nil
}

func (m *MockStripeService) GetCapability(ctx context.Context, capabilityID string) (*stripe.Capability, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	capability := &stripe.Capability{
		ID:   capabilityID,
		Key:  "card_payments",
	}
	
	return capability, nil
}

func (m *MockStripeService) UpdateCapability(ctx context.Context, capabilityID string, params *stripe.CapabilityParams) (*stripe.Capability, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	capability := &stripe.Capability{
		ID:   capabilityID,
		Key:  params.Key,
	}
	
	return capability, nil
}

func (m *MockStripeService) DeleteCapability(ctx context.Context, capabilityID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListCapabilities(ctx context.Context, params *stripe.CapabilityListParams) (*stripe.CapabilityList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	capabilities := []*stripe.Capability{
		{
			ID:   "cap_mock_1",
			Key:  "card_payments",
		},
		{
			ID:   "cap_mock_2",
			Key:  "transfers",
		},
	}
	
	return &stripe.CapabilityList{
		Data: capabilities,
	}, nil
}

func (m *MockStripeService) CreateCountrySpec(ctx context.Context, params *stripe.CountrySpecParams) (*stripe.CountrySpec, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	countrySpec := &stripe.CountrySpec{
		ID:            "country_mock",
		DefaultCurrency: &stripe.Currency{USD: true},
		SupportedPaymentMethods: []*stripe.PaymentMethod{
			{Type: "card"},
			{Type: "sepa_debit"},
		},
	}
	
	return countrySpec, nil
}

func (m *MockStripeService) GetCountrySpec(ctx context.Context, countrySpecID string) (*stripe.CountrySpec, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	countrySpec := &stripe.CountrySpec{
		ID:            countrySpecID,
		DefaultCurrency: &stripe.Currency{USD: true},
		SupportedPaymentMethods: []*stripe.PaymentMethod{
			{Type: "card"},
			{Type: "sepa_debit"},
		},
	}
	
	return countrySpec, nil
}

func (m *MockStripeService) UpdateCountrySpec(ctx context.Context, countrySpecID string, params *stripe.CountrySpecParams) (*stripe.CountrySpec, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	countrySpec := &stripe.CountrySpec{
		ID:            countrySpecID,
		DefaultCurrency: &stripe.Currency{USD: true},
		SupportedPaymentMethods: []*stripe.PaymentMethod{
			{Type: "card"},
			{Type: "sepa_debit"},
		},
	}
	
	return countrySpec, nil
}

func (m *MockStripeService) DeleteCountrySpec(ctx context.Context, countrySpecID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListCountrySpecs(ctx context.Context, params *stripe.CountrySpecListParams) (*stripe.CountrySpecList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	countrySpecs := []*stripe.CountrySpec{
		{
			ID:            "country_mock_1",
			DefaultCurrency: &stripe.Currency{USD: true},
			SupportedPaymentMethods: []*stripe.PaymentMethod{
				{Type: "card"},
				{Type: "sepa_debit"},
			},
		},
		{
			ID:            "country_mock_2",
			DefaultCurrency: &stripe.Currency{EUR: true},
			SupportedPaymentMethods: []*stripe.PaymentMethod{
				{Type: "card"},
				{Type: "sepa_debit"},
			},
		},
	}
	
	return &stripe.CountrySpecList{
		Data: countrySpecs,
	}, nil
}

func (m *MockStripeService) CreateCurrency(ctx context.Context, params *stripe.CurrencyParams) (*stripe.Currency, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	currency := &stripe.Currency{
		USD: true,
	}
	
	return currency, nil
}

func (m *MockStripeService) GetCurrency(ctx context.Context, currencyID string) (*stripe.Currency, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	currency := &stripe.Currency{
		USD: true,
	}
	
	return currency, nil
}

func (m *MockStripeService) UpdateCurrency(ctx context.Context, currencyID string, params *stripe.CurrencyParams) (*stripe.Currency, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	currency := &stripe.Currency{
		USD: true,
	}
	
	return currency, nil
}

func (m *MockStripeService) DeleteCurrency(ctx context.Context, currencyID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListCurrencies(ctx context.Context, params *stripe.CurrencyListParams) (*stripe.CurrencyList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	currencies := []*stripe.Currency{
		{USD: true},
		{EUR: true},
		{GBP: true},
	}
	
	return &stripe.CurrencyList{
		Data: currencies,
	}, nil
}

func (m *MockStripeService) CreateExchangeRate(ctx context.Context, params *stripe.ExchangeRateParams) (*stripe.ExchangeRate, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	exchangeRate := &stripe.ExchangeRate{
		ID:     "rate_mock",
		Rate:   1.2,
	}
	
	return exchangeRate, nil
}

func (m *MockStripeService) GetExchangeRate(ctx context.Context, exchangeRateID string) (*stripe.ExchangeRate, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	exchangeRate := &stripe.ExchangeRate{
		ID:     exchangeRateID,
		Rate:   1.2,
	}
	
	return exchangeRate, nil
}

func (m *MockStripeService) UpdateExchangeRate(ctx context.Context, exchangeRateID string, params *stripe.ExchangeRateParams) (*stripe.ExchangeRate, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	exchangeRate := &stripe.ExchangeRate{
		ID:     exchangeRateID,
		Rate:   params.Rate,
	}
	
	return exchangeRate, nil
}

func (m *MockStripeService) DeleteExchangeRate(ctx context.Context, exchangeRateID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListExchangeRates(ctx context.Context, params *stripe.ExchangeRateListParams) (*stripe.ExchangeRateList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	exchangeRates := []*stripe.ExchangeRate{
		{
			ID:   "rate_mock_1",
			Rate:  1.2,
		},
		{
			ID:   "rate_mock_2",
			Rate:  0.9,
		},
	}
	
	return &stripe.ExchangeRateList{
		Data: exchangeRates,
	}, nil
}

func (m *MockStripeService) CreateUsageRecord(ctx context.Context, params *stripe.UsageRecordParams) (*stripe.UsageRecord, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	usageRecord := &stripe.UsageRecord{
		ID:     "usage_mock",
		Quantity: 100,
	}
	
	return usageRecord, nil
}

func (m *MockStripeService) GetUsageRecord(ctx context.Context, usageRecordID string) (*stripe.UsageRecord, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	usageRecord := &stripe.UsageRecord{
		ID:     usageRecordID,
		Quantity: 100,
	}
	
	return usageRecord, nil
}

func (m *MockStripeService) UpdateUsageRecord(ctx context.Context, usageRecordID string, params *stripe.UsageRecordParams) (*stripe.UsageRecord, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	usageRecord := &stripe.UsageRecord{
		ID:     usageRecordID,
		Quantity: params.Quantity,
	}
	
	return usageRecord, nil
}

func (m *MockStripeService) DeleteUsageRecord(ctx context.Context, usageRecordID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListUsageRecords(ctx context.Context, params *stripe.UsageRecordListParams) (*stripe.UsageRecordList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	usageRecords := []*stripe.UsageRecord{
		{
			ID:     "usage_mock_1",
			Quantity: 100,
		},
		{
			ID:     "usage_mock_2",
			Quantity: 200,
		},
	}
	
	return &stripe.UsageRecordList{
		Data: usageRecords,
	}, nil
}

func (m *MockStripeService) CreateUsageRecordSummary(ctx context.Context, params *stripe.UsageRecordSummaryParams) (*stripe.UsageRecordSummary, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	summary := &stripe.UsageRecordSummary{
		ID:     "summary_mock",
		Period: "monthly",
	}
	
	return summary, nil
}

func (m *MockStripeService) GetUsageRecordSummary(ctx context.Context, summaryID string) (*stripe.UsageRecordSummary, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	summary := &stripe.UsageRecordSummary{
		ID:     summaryID,
		Period: "monthly",
	}
	
	return summary, nil
}

func (m *MockStripeService) UpdateUsageRecordSummary(ctx context.Context, summaryID string, params *stripe.UsageRecordSummaryParams) (*stripe.UsageRecordSummary, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	summary := &stripe.UsageRecordSummary{
		ID:     summaryID,
		Period: params.Period,
	}
	
	return summary, nil
}

func (m *MockStripeService) DeleteUsageRecordSummary(ctx context.Context, summaryID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListUsageRecordSummaries(ctx context.Context, params *stripe.UsageRecordSummaryListParams) (*stripe.UsageRecordSummaryList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	summaries := []*stripe.UsageRecordSummary{
		{
			ID:     "summary_mock_1",
			Period: "monthly",
		},
		{
			ID:     "summary_mock_2",
			Period: "yearly",
		},
	}
	
	return &stripe.UsageRecordSummaryList{
		Data: summaries,
	}, nil
}

func (m *MockStripeService) CreateScheduledQueryRun(ctx context.Context, params *stripe.ScheduledQueryRunParams) (*stripe.ScheduledQueryRun, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	run := &stripe.ScheduledQueryRun{
		ID:     "query_run_mock",
		Status: "completed",
	}
	
	return run, nil
}

func (m *MockStripeService) GetScheduledQueryRun(ctx context.Context, runID string) (*stripe.ScheduledQueryRun, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	run := &stripe.ScheduledQueryRun{
		ID:     runID,
		Status: "completed",
	}
	
	return run, nil
}

func (m *MockStripeService) UpdateScheduledQueryRun(ctx context.Context, runID string, params *stripe.ScheduledQueryRunParams) (*stripe.ScheduledQueryRun, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	run := &stripe.ScheduledQueryRun{
		ID:     runID,
		Status: params.Status,
	}
	
	return run, nil
}

func (m *MockStripeService) DeleteScheduledQueryRun(ctx context.Context, runID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListScheduledQueryRuns(ctx context.Context, params *stripe.ScheduledQueryRunListParams) (*stripe.ScheduledQueryRunList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	runs := []*stripe.ScheduledQueryRun{
		{
			ID:     "query_run_mock_1",
			Status: "completed",
		},
		{
			ID:     "query_run_mock_2",
			Status: "failed",
		},
	}
	
	return &stripe.ScheduledQueryRunList{
		Data: runs,
	}, nil
}

func (m *MockStripeService) CreateScheduledQuery(ctx context.Context, params *stripe.ScheduledQueryParams) (*stripe.ScheduledQuery, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	query := &stripe.ScheduledQuery{
		ID:     "query_mock",
		Status: "active",
	}
	
	return query, nil
}

func (m *MockStripeService) GetScheduledQuery(ctx context.Context, queryID string) (*stripe.ScheduledQuery, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	query := &stripe.ScheduledQuery{
		ID:     queryID,
		Status: "active",
	}
	
	return query, nil
}

func (m *MockStripeService) UpdateScheduledQuery(ctx context.Context, queryID string, params *stripe.ScheduledQueryParams) (*stripe.ScheduledQuery, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	query := &stripe.ScheduledQuery{
		ID:     queryID,
		Status: params.Status,
	}
	
	return query, nil
}

func (m *MockStripeService) DeleteScheduledQuery(ctx context.Context, queryID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListScheduledQueries(ctx context.Context, params *stripe.ScheduledQueryListParams) (*stripe.ScheduledQueryList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	queries := []*stripe.ScheduledQuery{
		{
			ID:     "query_mock_1",
			Status: "active",
		},
		{
			ID:     "query_mock_2",
			Status: "completed",
		},
	}
	
	return &stripe.ScheduledQueryList{
		Data: queries,
	}, nil
}

func (m *MockStripeService) CreateWebhookEndpoint(ctx context.Context, params *stripe.WebhookEndpointParams) (*stripe.WebhookEndpoint, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	endpoint := &stripe.WebhookEndpoint{
		ID:     "endpoint_mock",
		URL:    "https://example.com/webhook",
		Status: "enabled",
	}
	
	return endpoint, nil
}

func (m *MockStripeService) GetWebhookEndpoint(ctx context.Context, endpointID string) (*stripe.WebhookEndpoint, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	endpoint := &stripe.WebhookEndpoint{
		ID:     endpointID,
		URL:    "https://example.com/webhook",
		Status: "enabled",
	}
	
	return endpoint, nil
}

func (m *MockStripeService) UpdateWebhookEndpoint(ctx context.Context, endpointID string, params *stripe.WebhookEndpointParams) (*stripe.WebhookEndpoint, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	endpoint := &stripe.WebhookEndpoint{
		ID:     endpointID,
		URL:    params.URL,
		Status: params.Status,
	}
	
	return endpoint, nil
}

func (m *MockStripeService) DeleteWebhookEndpoint(ctx context.Context, endpointID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListWebhookEndpoints(ctx context.Context, params *stripe.WebhookEndpointListParams) (*stripe.WebhookEndpointList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	endpoints := []*stripe.WebhookEndpoint{
		{
			ID:     "endpoint_mock_1",
			URL:    "https://example.com/webhook1",
			Status: "enabled",
		},
		{
			ID:     "endpoint_mock_2",
			URL:    "https://example.com/webhook2",
			Status: "disabled",
		},
	}
	
	return &stripe.WebhookEndpointList{
		Data: endpoints,
	}, nil
}

func (m *MockStripeService) CreateEvent(ctx context.Context, params *stripe.EventParams) (*stripe.Event, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	event := &stripe.Event{
		ID:   "event_mock",
		Type: "payment_intent.succeeded",
	}
	
	return event, nil
}

func (m *MockStripeService) GetEvent(ctx context.Context, eventID string) (*stripe.Event, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	event := &stripe.Event{
		ID:   eventID,
		Type: "payment_intent.succeeded",
	}
	
	return event, nil
}

func (m *MockStripeService) ListEvents(ctx context.Context, params *stripe.EventListParams) (*stripe.EventList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	events := []*stripe.Event{
		{
			ID:   "event_mock_1",
			Type: "payment_intent.succeeded",
		},
		{
			ID:   "event_mock_2",
			Type: "invoice.payment_succeeded",
		},
	}
	
	return &stripe.EventList{
		Data: events,
	}, nil
}

func (m *MockStripeService) CreateEarlyFraudWarning(ctx context.Context, params *stripe.EarlyFraudWarningParams) (*stripe.EarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.EarlyFraudWarning{
		ID:     "warning_mock",
		Action: "review",
	}
	
	return warning, nil
}

func (m *MockStripeService) GetEarlyFraudWarning(ctx context.Context, warningID string) (*stripe.EarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.EarlyFraudWarning{
		ID:     warningID,
		Action: "review",
	}
	
	return warning, nil
}

func (m *MockStripeService) UpdateEarlyFraudWarning(ctx context.Context, warningID string, params *stripe.EarlyFraudWarningParams) (*stripe.EarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.EarlyFraudWarning{
		ID:     warningID,
		Action: params.Action,
	}
	
	return warning, nil
}

func (m *MockStripeService) DeleteEarlyFraudWarning(ctx context.Context, warningID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListEarlyFraudWarnings(ctx context.Context, params *stripe.EarlyFraudWarningListParams) (*stripe.EarlyFraudWarningList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warnings := []*stripe.EarlyFraudWarning{
		{
			ID:     "warning_mock_1",
			Action: "review",
		},
		{
			ID:     "warning_mock_2",
			Action: "block",
		},
	}
	
	return &stripe.EarlyFraudWarningList{
		Data: warnings,
	}, nil
}

func (m *MockStripeService) CreateRadarValueList(ctx context.Context, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     "value_list_mock",
		Name:   "Mock Value List",
	}
	
	return valueList, nil
}

func (m *MockStripeService) GetRadarValueList(ctx context.Context, valueListID string) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     valueListID,
		Name:   "Mock Value List",
	}
	
	return valueList, nil
}

func (m *MockStripeService) UpdateRadarValueList(ctx context.Context, valueListID string, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     valueListID,
		Name:   params.Name,
	}
	
	return valueList, nil
}

func (m *MockStripeService) DeleteRadarValueList(ctx context.Context, valueListID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarValueLists(ctx context.Context, params *stripe.RadarValueListListParams) (*stripe.RadarValueListList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueLists := []*stripe.RadarValueList{
		{
			ID:   "value_list_mock_1",
			Name: "Mock Value List 1",
		},
		{
			ID:   "value_list_mock_2",
			Name: "Mock Value List 2",
		},
	}
	
	return &stripe.RadarValueListList{
		Data: valueLists,
	}, nil
}

func (m *MockStripeService) CreateRadarRule(ctx context.Context, params *stripe.RadarRuleParams) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     "rule_mock",
		Name:   "Mock Rule",
	}
	
	return rule, nil
}

func (m *MockStripeService) GetRadarRule(ctx context.Context, ruleID string) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     ruleID,
		Name:   "Mock Rule",
	}
	
	return rule, nil
}

func (m *MockStripeService) UpdateRadarRule(ctx context.Context, ruleID string, params *stripe.RadarRuleParams) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     ruleID,
		Name:   params.Name,
	}
	
	return rule, nil
}

func (m *MockStripeService) DeleteRadarRule(ctx context.Context, ruleID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarRules(ctx context.Context, params *stripe.RadarRuleListParams) (*stripe.RadarRuleList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rules := []*stripe.RadarRule{
		{
			ID:   "rule_mock_1",
			Name: "Mock Rule 1",
		},
		{
			ID:   "rule_mock_2",
			Name: "Mock Rule 2",
		},
	}
	
	return &stripe.RadarRuleList{
		Data: rules,
	}, nil
}

func (m *MockStripeService) CreateRadarValueListItem(ctx context.Context, params *stripe.RadarValueListItemParams) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    "item_mock",
		Value: "test_value",
	}
	
	return item, nil
}

func (m *MockStripeService) GetRadarValueListItem(ctx context.Context, itemID string) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    itemID,
		Value: "test_value",
	}
	
	return item, nil
}

func (m *MockStripeService) UpdateRadarValueListItem(ctx context.Context, itemID string, params *stripe.RadarValueListItemParams) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    itemID,
		Value: params.Value,
	}
	
	return item, nil
}

func (m *MockStripeService) DeleteRadarValueListItem(ctx context.Context, itemID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarValueListItems(ctx context.Context, params *stripe.RadarValueListItemListParams) (*stripe.RadarValueListItemList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	items := []*stripe.RadarValueListItem{
		{
			ID:    "item_mock_1",
			Value: "test_value_1",
		},
		{
			ID:    "item_mock_2",
			Value: "test_value_2",
		},
	}
	
	return &stripe.RadarValueListItemList{
		Data: items,
	}, nil
}

func (m *MockStripeService) CreateRadarReview(ctx context.Context, params *stripe.RadarReviewParams) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     "review_mock",
		Reason: "unusual_activity",
	}
	
	return review, nil
}

func (m *MockStripeService) GetRadarReview(ctx context.Context, reviewID string) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     reviewID,
		Reason: "unusual_activity",
	}
	
	return review, nil
}

func (m *MockStripeService) UpdateRadarReview(ctx context.Context, reviewID string, params *stripe.RadarReviewParams) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     reviewID,
		Reason: params.Reason,
	}
	
	return review, nil
}

func (m *MockStripeService) DeleteRadarReview(ctx context.Context, reviewID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarReviews(ctx context.Context, params *stripe.RadarReviewListParams) (*stripe.RadarReviewList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	reviews := []*stripe.RadarReview{
		{
			ID:     "review_mock_1",
			Reason: "unusual_activity",
		},
		{
			ID:     "review_mock_2",
			Reason: "high_risk",
		},
	}
	
	return &stripe.RadarReviewList{
		Data: reviews,
	}, nil
}

func (m *MockStripeService) CreateRadarReviewSession(ctx context.Context, params *stripe.RadarReviewSessionParams) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     "session_mock",
		Redacted: false,
	}
	
	return session, nil
}

func (m *MockStripeService) GetRadarReviewSession(ctx context.Context, sessionID string) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     sessionID,
		Redacted: false,
	}
	
	return session, nil
}

func (m *MockStripeService) UpdateRadarReviewSession(ctx context.Context, sessionID string, params *stripe.RadarReviewSessionParams) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     sessionID,
		Redacted: params.Redacted,
	}
	
	return session, nil
}

func (m *MockStripeService) DeleteRadarReviewSession(ctx context.Context, sessionID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarReviewSessions(ctx context.Context, params *stripe.RadarReviewSessionListParams) (*stripe.RadarReviewSessionList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	sessions := []*stripe.RadarReviewSession{
		{
			ID:     "session_mock_1",
			Redacted: false,
		},
		{
			ID:     "session_mock_2",
			Redacted: true,
		},
	}
	
	return &stripe.RadarReviewSessionList{
		Data: sessions,
	}, nil
}

func (m *MockStripeService) CreateRadarEarlyFraudWarning(ctx context.Context, params *stripe.RadarEarlyFraudWarningParams) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     "warning_mock",
		Action: "review",
	}
	
	return warning, nil
}

func (m *MockStripeService) GetRadarEarlyFraudWarning(ctx context.Context, warningID string) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     warningID,
		Action: "review",
	}
	
	return warning, nil
}

func (m *MockStripeService) UpdateRadarEarlyFraudWarning(ctx context.Context, warningID string, params *stripe.RadarEarlyFraudWarningParams) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     warningID,
		Action: params.Action,
	}
	
	return warning, nil
}

func (m *MockStripeService) DeleteRadarEarlyFraudWarning(ctx context.Context, warningID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarEarlyFraudWarnings(ctx context.Context, params *stripe.RadarEarlyFraudWarningListParams) (*stripe.RadarEarlyFraudWarningList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warnings := []*stripe.RadarEarlyFraudWarning{
		{
			ID:     "warning_mock_1",
			Action: "review",
		},
		{
			ID:     "warning_mock_2",
			Action: "block",
		},
	}
	
	return &stripe.RadarEarlyFraudWarningList{
		Data: warnings,
	}, nil
}

func (m *MockStripeService) CreateRadarValueList(ctx context.Context, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     "value_list_mock",
		Name:   "Mock Value List",
	}
	
	return valueList, nil
}

func (m *MockStripeService) GetRadarValueList(ctx context.Context, valueListID string) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     valueListID,
		Name:   "Mock Value List",
	}
	
	return valueList, nil
}

func (m *MockStripeService) UpdateRadarValueList(ctx context.Context, valueListID string, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     valueListID,
		Name:   params.Name,
	}
	
	return valueList, nil
}

func (m *MockStripeService) DeleteRadarValueList(ctx context.Context, valueListID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarValueLists(ctx context.Context, params *stripe.RadarValueListListParams) (*stripe.RadarValueListList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueLists := []*stripe.RadarValueList{
		{
			ID:   "value_list_mock_1",
			Name: "Mock Value List 1",
		},
		{
			ID:   "value_list_mock_2",
			Name: "Mock Value List 2",
		},
	}
	
	return &stripe.RadarValueListList{
		Data: valueLists,
	}, nil
}

func (m *MockStripeService) CreateRadarRule(ctx context.Context, params *stripe.RadarRuleParams) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     "rule_mock",
		Name:   "Mock Rule",
	}
	
	return rule, nil
}

func (m *MockStripeService) GetRadarRule(ctx context.Context, ruleID string) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     ruleID,
		Name:   "Mock Rule",
	}
	
	return rule, nil
}

func (m *MockStripeService) UpdateRadarRule(ctx context.Context, ruleID string, params *stripe.RadarRuleParams) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     ruleID,
		Name:   params.Name,
	}
	
	return rule, nil
}

func (m *MockStripeService) DeleteRadarRule(ctx context.Context, ruleID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarRules(ctx context.Context, params *stripe.RadarRuleListParams) (*stripe.RadarRuleList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rules := []*stripe.RadarRule{
		{
			ID:   "rule_mock_1",
			Name: "Mock Rule 1",
		},
		{
			ID:   "rule_mock_2",
			Name: "Mock Rule 2",
		},
	}
	
	return &stripe.RadarRuleList{
		Data: rules,
	}, nil
}

func (m *MockStripeService) CreateRadarValueListItem(ctx context.Context, params *stripe.RadarValueListItemParams) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    "item_mock",
		Value: "test_value",
	}
	
	return item, nil
}

func (m *MockStripeService) GetRadarValueListItem(ctx context.Context, itemID string) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    itemID,
		Value: "test_value",
	}
	
	return item, nil
}

func (m *MockStripeService) UpdateRadarValueListItem(ctx context.Context, itemID string, params *stripe.RadarValueListItemParams) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    itemID,
		Value: params.Value,
	}
	
	return item, nil
}

func (m *MockStripeService) DeleteRadarValueListItem(ctx context.Context, itemID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarValueListItems(ctx context.Context, params *stripe.RadarValueListItemListParams) (*stripe.RadarValueListItemList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	items := []*stripe.RadarValueListItem{
		{
			ID:    "item_mock_1",
			Value: "test_value_1",
		},
		{
			ID:    "item_mock_2",
			Value: "test_value_2",
		},
	}
	
	return &stripe.RadarValueListItemList{
		Data: items,
	}, nil
}

func (m *MockStripeService) CreateRadarReview(ctx context.Context, params *stripe.RadarReviewParams) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     "review_mock",
		Reason: "unusual_activity",
	}
	
	return review, nil
}

func (m *MockStripeService) GetRadarReview(ctx context.Context, reviewID string) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     reviewID,
		Reason: "unusual_activity",
	}
	
	return review, nil
}

func (m *MockStripeService) UpdateRadarReview(ctx context.Context, reviewID string, params *stripe.RadarReviewParams) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     reviewID,
		Reason: params.Reason,
	}
	
	return review, nil
}

func (m *MockStripeService) DeleteRadarReview(ctx context.Context, reviewID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarReviews(ctx context.Context, params *stripe.RadarReviewListParams) (*stripe.RadarReviewList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	reviews := []*stripe.RadarReview{
		{
			ID:     "review_mock_1",
			Reason: "unusual_activity",
		},
		{
			ID:     "review_mock_2",
			Reason: "high_risk",
		},
	}
	
	return &stripe.RadarReviewList{
		Data: reviews,
	}, nil
}

func (m *MockStripeService) CreateRadarReviewSession(ctx context.Context, params *stripe.RadarReviewSessionParams) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     "session_mock",
		Redacted: false,
	}
	
	return session, nil
}

func (m *MockStripeService) GetRadarReviewSession(ctx context.Context, sessionID string) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     sessionID,
		Redacted: false,
	}
	
	return session, nil
}

func (m *MockStripeService) UpdateRadarReviewSession(ctx context.Context, sessionID string, params *stripe.RadarReviewSessionParams) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     sessionID,
		Redacted: params.Redacted,
	}
	
	return session, nil
}

func (m *MockStripeService) DeleteRadarReviewSession(ctx context.Context, sessionID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarReviewSessions(ctx context.Context, params *stripe.RadarReviewSessionListParams) (*stripe.RadarReviewSessionList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	sessions := []*stripe.RadarReviewSession{
		{
			ID:     "session_mock_1",
			Redacted: false,
		},
		{
			ID:     "session_mock_2",
			Redacted: true,
		},
	}
	
	return &stripe.RadarReviewSessionList{
		Data: sessions,
	}, nil
}

func (m *MockStripeService) CreateRadarEarlyFraudWarning(ctx context.Context, params *stripe.RadarEarlyFraudWarningParams) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     "warning_mock",
		Action: "review",
	}
	
	return warning, nil
}

func (m *MockStripeService) GetRadarEarlyFraudWarning(ctx context.Context, warningID string) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     warningID,
		Action: "review",
	}
	
	return warning, nil
}

func (m *MockStripeService) UpdateRadarEarlyFraudWarning(ctx context.Context, warningID string, params *stripe.RadarEarlyFraudWarningParams) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     warningID,
		Action: params.Action,
	}
	
	return warning, nil
}

func (m *MockStripeService) DeleteRadarEarlyFraudWarning(ctx context.Context, warningID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarEarlyFraudWarnings(ctx context.Context, params *stripe.RadarEarlyFraudWarningListParams) (*stripe.RadarEarlyFraudWarningList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warnings := []*stripe.RadarEarlyFraudWarning{
		{
			ID:     "warning_mock_1",
			Action: "review",
		},
		{
			ID:     "warning_mock_2",
			Action: "block",
		},
	}
	
	return &stripe.RadarEarlyFraudWarningList{
		Data: warnings,
	}, nil
}

func (m *MockStripeService) CreateRadarValueList(ctx context.Context, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     "value_list_mock",
		Name:   "Mock Value List",
	}
	
	return valueList, nil
}

func (m *MockStripeService) GetRadarValueList(ctx context.Context, valueListID string) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     valueListID,
		Name:   "Mock Value List",
	}
	
	return valueList, nil
}

func (m *MockStripeService) UpdateRadarValueList(ctx context.Context, valueListID string, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     valueListID,
		Name:   params.Name,
	}
	
	return valueList, nil
}

func (m *MockStripeService) DeleteRadarValueList(ctx context.Context, valueListID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarValueLists(ctx context.Context, params *stripe.RadarValueListListParams) (*stripe.RadarValueListList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueLists := []*stripe.RadarValueList{
		{
			ID:   "value_list_mock_1",
			Name: "Mock Value List 1",
		},
		{
			ID:   "value_list_mock_2",
			Name: "Mock Value List 2",
		},
	}
	
	return &stripe.RadarValueListList{
		Data: valueLists,
	}, nil
}

func (m *MockStripeService) CreateRadarRule(ctx context.Context, params *stripe.RadarRuleParams) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     "rule_mock",
		Name:   "Mock Rule",
	}
	
	return rule, nil
}

func (m *MockStripeService) GetRadarRule(ctx context.Context, ruleID string) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     ruleID,
		Name:   "Mock Rule",
	}
	
	return rule, nil
}

func (m *MockStripeService) UpdateRadarRule(ctx context.Context, ruleID string, params *stripe.RadarRuleParams) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     ruleID,
		Name:   params.Name,
	}
	
	return rule, nil
}

func (m *MockStripeService) DeleteRadarRule(ctx context.Context, ruleID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarRules(ctx context.Context, params *stripe.RadarRuleListParams) (*stripe.RadarRuleList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rules := []*stripe.RadarRule{
		{
			ID:   "rule_mock_1",
			Name: "Mock Rule 1",
		},
		{
			ID:   "rule_mock_2",
			Name: "Mock Rule 2",
		},
	}
	
	return &stripe.RadarRuleList{
		Data: rules,
	}, nil
}

func (m *MockStripeService) CreateRadarValueListItem(ctx context.Context, params *stripe.RadarValueListItemParams) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    "item_mock",
		Value: "test_value",
	}
	
	return item, nil
}

func (m *MockStripeService) GetRadarValueListItem(ctx context.Context, itemID string) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    itemID,
		Value: "test_value",
	}
	
	return item, nil
}

func (m *MockStripeService) UpdateRadarValueListItem(ctx context.Context, itemID string, params *stripe.RadarValueListItemParams) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    itemID,
		Value: params.Value,
	}
	
	return item, nil
}

func (m *MockStripeService) DeleteRadarValueListItem(ctx context.Context, itemID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarValueListItems(ctx context.Context, params *stripe.RadarValueListItemListParams) (*stripe.RadarValueListItemList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	items := []*stripe.RadarValueListItem{
		{
			ID:    "item_mock_1",
			Value: "test_value_1",
		},
		{
			ID:    "item_mock_2",
			Value: "test_value_2",
		},
	}
	
	return &stripe.RadarValueListItemList{
		Data: items,
	}, nil
}

func (m *MockStripeService) CreateRadarReview(ctx context.Context, params *stripe.RadarReviewParams) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     "review_mock",
		Reason: "unusual_activity",
	}
	
	return review, nil
}

func (m *MockStripeService) GetRadarReview(ctx context.Context, reviewID string) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     reviewID,
		Reason: "unusual_activity",
	}
	
	return review, nil
}

func (m *MockStripeService) UpdateRadarReview(ctx context.Context, reviewID string, params *stripe.RadarReviewParams) (*stripe.RadarReview, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	review := &stripe.RadarReview{
		ID:     reviewID,
		Reason: params.Reason,
	}
	
	return review, nil
}

func (m *MockStripeService) DeleteRadarReview(ctx context.Context, reviewID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarReviews(ctx context.Context, params *stripe.RadarReviewListParams) (*stripe.RadarReviewList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	reviews := []*stripe.RadarReview{
		{
			ID:     "review_mock_1",
			Reason: "unusual_activity",
		},
		{
			ID:     "review_mock_2",
			Reason: "high_risk",
		},
	}
	
	return &stripe.RadarReviewList{
		Data: reviews,
	}, nil
}

func (m *MockStripeService) CreateRadarReviewSession(ctx context.Context, params *stripe.RadarReviewSessionParams) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     "session_mock",
		Redacted: false,
	}
	
	return session, nil
}

func (m *MockStripeService) GetRadarReviewSession(ctx context.Context, sessionID string) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     sessionID,
		Redacted: false,
	}
	
	return session, nil
}

func (m *MockStripeService) UpdateRadarReviewSession(ctx context.Context, sessionID string, params *stripe.RadarReviewSessionParams) (*stripe.RadarReviewSession, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	session := &stripe.RadarReviewSession{
		ID:     sessionID,
		Redacted: params.Redacted,
	}
	
	return session, nil
}

func (m *MockStripeService) DeleteRadarReviewSession(ctx context.Context, sessionID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarReviewSessions(ctx context.Context, params *stripe.RadarReviewSessionListParams) (*stripe.RadarReviewSessionList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	sessions := []*stripe.RadarReviewSession{
		{
			ID:     "session_mock_1",
			Redacted: false,
		},
		{
			ID:     "session_mock_2",
			Redacted: true,
		},
	}
	
	return &stripe.RadarReviewSessionList{
		Data: sessions,
	}, nil
}

func (m *MockStripeService) CreateRadarEarlyFraudWarning(ctx context.Context, params *stripe.RadarEarlyFraudWarningParams) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     "warning_mock",
		Action: "review",
	}
	
	return warning, nil
}

func (m *MockStripeService) GetRadarEarlyFraudWarning(ctx context.Context, warningID string) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     warningID,
		Action: "review",
	}
	
	return warning, nil
}

func (m *MockStripeService) UpdateRadarEarlyFraudWarning(ctx context.Context, warningID string, params *stripe.RadarEarlyFraudWarningParams) (*stripe.RadarEarlyFraudWarning, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warning := &stripe.RadarEarlyFraudWarning{
		ID:     warningID,
		Action: params.Action,
	}
	
	return warning, nil
}

func (m *MockStripeService) DeleteRadarEarlyFraudWarning(ctx context.Context, warningID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarEarlyFraudWarnings(ctx context.Context, params *stripe.RadarEarlyFraudWarningListParams) (*stripe.RadarEarlyFraudWarningList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	warnings := []*stripe.RadarEarlyFraudWarning{
		{
			ID:     "warning_mock_1",
			Action: "review",
		},
		{
			ID:     "warning_mock_2",
			Action: "block",
		},
	}
	
	return &stripe.RadarEarlyFraudWarningList{
		Data: warnings,
	}, nil
}

func (m *MockStripeService) CreateRadarValueList(ctx context.Context, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     "value_list_mock",
		Name:   "Mock Value List",
	}
	
	return valueList, nil
}

func (m *MockStripeService) GetRadarValueList(ctx context.Context, valueListID string) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     valueListID,
		Name:   "Mock Value List",
	}
	
	return valueList, nil
}

func (m *MockStripeService) UpdateRadarValueList(ctx context.Context, valueListID string, params *stripe.RadarValueListParams) (*stripe.RadarValueList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueList := &stripe.RadarValueList{
		ID:     valueListID,
		Name:   params.Name,
	}
	
	return valueList, nil
}

func (m *MockStripeService) DeleteRadarValueList(ctx context.Context, valueListID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarValueLists(ctx context.Context, params *stripe.RadarValueListListParams) (*stripe.RadarValueListList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	valueLists := []*stripe.RadarValueList{
		{
			ID:   "value_list_mock_1",
			Name: "Mock Value List 1",
		},
		{
			ID:   "value_list_mock_2",
			Name: "Mock Value List 2",
		},
	}
	
	return &stripe.RadarValueListList{
		Data: valueLists,
	}, nil
}

func (m *MockStripeService) CreateRadarRule(ctx context.Context, params *stripe.RadarRuleParams) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     "rule_mock",
		Name:   "Mock Rule",
	}
	
	return rule, nil
}

func (m *MockStripeService) GetRadarRule(ctx context.Context, ruleID string) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     ruleID,
		Name:   "Mock Rule",
	}
	
	return rule, nil
}

func (m *MockStripeService) UpdateRadarRule(ctx context.Context, ruleID string, params *stripe.RadarRuleParams) (*stripe.RadarRule, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rule := &stripe.RadarRule{
		ID:     ruleID,
		Name:   params.Name,
	}
	
	return rule, nil
}

func (m *MockStripeService) DeleteRadarRule(ctx context.Context, ruleID string) error {
	if m.simulateError {
		return &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	return nil
}

func (m *MockStripeService) ListRadarRules(ctx context.Context, params *stripe.RadarRuleListParams) (*stripe.RadarRuleList, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	rules := []*stripe.RadarRule{
		{
			ID:   "rule_mock_1",
			Name: "Mock Rule 1",
		},
		{
			ID:   "rule_mock_2",
			Name: "Mock Rule 2",
		},
	}
	
	return &stripe.RadarRuleList{
		Data: rules,
	}, nil
}

func (m *MockStripeService) CreateRadarValueListItem(ctx context.Context, params *stripe.RadarValueListItemParams) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    "item_mock",
		Value: "test_value",
	}
	
	return item, nil
}

func (m *MockStripeService) GetRadarValueListItem(ctx context.Context, itemID string) (*stripe.RadarValueListItem, error) {
	if m.simulateError {
		return nil, &stripe.Error{Msg: m.errorMessage}
	}
	
	// Mock implementation
	item := &stripe.RadarValueListItem{
		ID:    itemID,
		Value: "test_value",
	}
	
	return item, nil
}

func (m *MockStripeService) Upd