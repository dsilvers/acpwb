PTO_DAYS_BY_YEAR = {
    range(1993, 2000): 10,
    range(2000, 2008): 12,
    range(2008, 2016): 15,
    range(2016, 2021): 18,
    range(2021, 2026): 20,
}

SICK_DAYS_BY_YEAR = {
    range(1993, 2005): 5,
    range(2005, 2016): 7,
    range(2016, 2026): 10,
}

EXPENSE_APPROVAL_THRESHOLD_BY_YEAR = {
    range(1993, 2000): 500,
    range(2000, 2008): 750,
    range(2008, 2016): 1000,
    range(2016, 2021): 1500,
    range(2021, 2026): 2000,
}

MEAL_PER_DIEM_BY_YEAR = {
    range(1993, 2000): 35,
    range(2000, 2008): 45,
    range(2008, 2016): 55,
    range(2016, 2021): 65,
    range(2021, 2026): 75,
}

HOTEL_PER_DIEM_BY_YEAR = {
    range(1993, 2000): 150,
    range(2000, 2008): 200,
    range(2008, 2016): 250,
    range(2016, 2021): 300,
    range(2021, 2026): 350,
}

BEREAVEMENT_DAYS = {
    'immediate': 5,
    'extended': 3,
    'other': 1,
}

PARENTAL_LEAVE_WEEKS_BY_YEAR = {
    range(1993, 2005): 6,
    range(2005, 2016): 8,
    range(2016, 2021): 12,
    range(2021, 2026): 16,
}

NOTICE_PERIODS_BY_YEAR = {
    range(1993, 2010): '2 weeks',
    range(2010, 2020): '30 days',
    range(2020, 2026): '30 days',
}

REMOTE_WORK_POLICY_BY_YEAR = {
    range(1993, 2020): 'not available except in limited circumstances approved by the Senior Vice President of Human Resources',
    range(2020, 2022): 'available during the declared public health emergency with manager approval; equipment reimbursement up to $800',
    range(2022, 2026): 'available for eligible positions with manager and HR approval; equipment reimbursement up to $1,200 per fiscal year',
}

TUITION_REIMBURSEMENT_BY_YEAR = {
    range(1993, 2005): 3000,
    range(2005, 2016): 4000,
    range(2016, 2021): 5250,
    range(2021, 2026): 5250,
}


def get_threshold(mapping, year):
    for year_range, value in mapping.items():
        if year in year_range:
            return value
    return list(mapping.values())[-1]
