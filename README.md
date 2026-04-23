# 💰 Finance Insights

> Open-source personal finance platform that transforms raw bank
> statements into meaningful insights.

![License](https://img.shields.io/badge/license-MIT-green)
![Backend](https://img.shields.io/badge/backend-Go%201.25-blue)
![Frontend](https://img.shields.io/badge/frontend-React%20%2B%20TypeScript-blue)
![Database](https://img.shields.io/badge/database-PostgreSQL-blue)
![Status](https://img.shields.io/badge/status-WIP-orange)

# 💰 Personal Finance Application

## 📌 Application Description

An application that allows users to upload bank statements, parse
transactions, organize them through tagging, and generate insights and
reports.

------------------------------------------------------------------------

## 🧰 Tech Stack

-   Backend: Go
-   Frontend: React + TypeScript + Tailwind CSS
-   Database: PostgreSQL
-   Infrastructure: Docker Compose

------------------------------------------------------------------------

## 🏗️ High Level Design

### 1. Bank Statement Upload

-   User selects bank (e.g. Piraeus) and uploads statement
-   Frontend shows processing animation
-   Backend:
    -   Stores raw file in `/year/`
    -   Renames file to `year-bank-startdate`

------------------------------------------------------------------------

### 2. Parser Detection & Parsing

#### Raw Transaction Processing

-   Detect bank
-   Detect statement format/version
-   Extract rows
-   Normalize:
    -   Dates
    -   Amounts
    -   Balance
    -   Currency
    -   Description
    -   Reference IDs

------------------------------------------------------------------------

### 3. Normalized Transactions

-   Unified structure across all banks

------------------------------------------------------------------------

### 4. User Correction & Tagging

-   Merchant identifier (e.g. efood, dei, ab-vasilopoulos)
-   Categorization rules:
    -   Primary tag
    -   Secondary tags
    -   Notes
    -   Custom title
    -   Split rules

------------------------------------------------------------------------

### 5. Summaries & Analytics

-   Weekly, Monthly, Annual summaries
-   Comparisons and visualizations

------------------------------------------------------------------------

## 📂 File Handling

-   Raw statements stored by year
-   Normalized JSON stored separately
-   Data persisted in database

------------------------------------------------------------------------

## 🗄️ Database Schema

### Users

-   id
-   name
-   username
-   password

### Accounts

-   id
-   user_id
-   bank_name
-   account_number
-   currency

### Transactions

-   id
-   account_id
-   date
-   bank_reference_number
-   justification
-   indicator
-   merchant_identifier
-   amount1
-   mcc_code
-   card_masked
-   reference
-   description
-   payment_method
-   direction (debit/credit)
-   amount
-   balance_after_transaction
-   statement_file_name

### Manual Adjustments

-   id
-   date
-   user_id
-   direction
-   payment_method

### Merchants

-   id
-   identifier_name
-   primary_tag
-   secondary_tags
-   default_title

### Transaction Extra

-   id
-   transaction_id
-   note
-   in_report
-   parser_name_version

### Tags

-   id
-   name
-   type (primary/secondary)

------------------------------------------------------------------------

## 🧠 UI Features

### Organizing Tab

-   Unrecognized merchants
-   Uncategorized transactions
-   Duplicate detection
-   Manual adjustments
-   Bulk tagging
-   Exclude large one-off expenses (e.g. car purchase)

------------------------------------------------------------------------

## 🏷️ Tags

### Primary Tags

-   Food
-   Transport
-   Home
-   Family
-   Personal
-   Health
-   Finance
-   Leisure
-   Travel
-   Income
-   Transfers
-   Miscellaneous
-   SpecialCase

### Secondary Tags (Examples)

-   groceries
-   eating out
-   fuel
-   parking
-   electricity
-   rent
-   pharmacy
-   haircut
-   netflix
-   toys
-   school

------------------------------------------------------------------------

## 🖥️ Frontend Menu

-   Import
-   Transactions
-   Merchants / Rules
-   Calendar
-   Reports

------------------------------------------------------------------------

## 📊 Reports

### Standard

-   Weekly
-   Monthly
-   Annual

### Advanced

-   Month vs previous month
-   Year-over-year comparison
-   Rolling 12-month average
-   Spending by category
-   Top merchants
-   Recurring payments
-   Unusual spending detection
-   Cash flow trend
-   Family vs personal spend
-   Fixed vs variable costs
-   Average daily spend
-   Weekend vs weekday spend

------------------------------------------------------------------------

## 🚀 Future Features

### Recurring Transaction Detection

Automatically detect: - Rent - Salary - Subscriptions - Utilities - Loan
payments - Insurance

