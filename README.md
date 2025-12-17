# merchant_summary_demo_backend
# endpoint 
http://localhost:8080/api/merchant/summary

# method
post

# input request
{
  "mid": ["000000000001","000000000002","000000000003"]
}


# output request
{
    "error_schema": {
        "error_code": "D000",
        "error_message": {
            "english": "Success",
            "indonesian": "Berhasil"
        }
    },
    "output_schema": {
        "merchant_ids": [
            "000000000001",
            "000000000002",
            "000000000003"
        ],
        "current_date": "2025-12-17",
        "today_total_amount": "77803622",
        "weekly_total_amount": "584236342",
        "monthly_total_amount": "2513784014"
    }
}