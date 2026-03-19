import random
import hashlib

FIRST_NAMES = [
    'James', 'Mary', 'John', 'Patricia', 'Robert', 'Jennifer', 'Michael', 'Linda',
    'William', 'Barbara', 'David', 'Elizabeth', 'Richard', 'Susan', 'Joseph', 'Jessica',
    'Thomas', 'Sarah', 'Charles', 'Karen', 'Christopher', 'Lisa', 'Daniel', 'Nancy',
    'Matthew', 'Betty', 'Anthony', 'Margaret', 'Mark', 'Sandra', 'Donald', 'Ashley',
    'Steven', 'Dorothy', 'Paul', 'Kimberly', 'Andrew', 'Emily', 'Joshua', 'Donna',
    'Kenneth', 'Michelle', 'Kevin', 'Carol', 'Brian', 'Amanda', 'George', 'Melissa',
    'Timothy', 'Deborah', 'Ronald', 'Stephanie', 'Edward', 'Rebecca', 'Jason', 'Sharon',
    'Jeffrey', 'Laura', 'Ryan', 'Cynthia', 'Jacob', 'Kathleen', 'Gary', 'Amy',
    'Nicholas', 'Angela', 'Eric', 'Shirley', 'Jonathan', 'Anna', 'Stephen', 'Brenda',
    'Larry', 'Pamela', 'Justin', 'Emma', 'Scott', 'Nicole', 'Brandon', 'Helen',
    'Benjamin', 'Samantha', 'Samuel', 'Katherine', 'Raymond', 'Christine', 'Gregory',
    'Debra', 'Frank', 'Rachel', 'Alexander', 'Carolyn', 'Patrick', 'Janet', 'Jack',
    'Catherine', 'Dennis', 'Maria', 'Jerry', 'Heather', 'Tyler', 'Diane', 'Aaron',
    'Julie', 'Jose', 'Joyce', 'Adam', 'Victoria', 'Nathan', 'Kelly', 'Henry',
    'Christina', 'Douglas', 'Lauren', 'Zachary', 'Joan', 'Peter', 'Evelyn', 'Kyle',
    'Olivia', 'Walter', 'Judith', 'Ethan', 'Megan', 'Jeremy', 'Cheryl', 'Harold',
    'Andrea', 'Terry', 'Hannah', 'Sean', 'Martha', 'Austin', 'Jacqueline', 'Carl',
    'Frances', 'Arthur', 'Gloria', 'Lawrence', 'Ann', 'Dylan', 'Teresa', 'Jesse',
    'Kathryn', 'Jordan', 'Sara', 'Bryan', 'Janice', 'Billy', 'Jean', 'Bruce',
    'Alice', 'Gabriel', 'Madison', 'Joe', 'Doris', 'Logan', 'Abigail', 'Alan',
    'Julia', 'Juan', 'Judy', 'Albert', 'Rose', 'Willie', 'Grace', 'Wayne',
    'Denise', 'Randy', 'Amber', 'Mason', 'Marilyn', 'Roy', 'Beverly', 'Ralph',
    'Danielle', 'Eugene', 'Theresa', 'Russell', 'Sophia', 'Bobby', 'Marie',
]

LAST_NAMES = [
    'Smith', 'Johnson', 'Williams', 'Brown', 'Jones', 'Garcia', 'Miller', 'Davis',
    'Rodriguez', 'Martinez', 'Hernandez', 'Lopez', 'Gonzalez', 'Wilson', 'Anderson',
    'Thomas', 'Taylor', 'Moore', 'Jackson', 'Martin', 'Lee', 'Perez', 'Thompson',
    'White', 'Harris', 'Sanchez', 'Clark', 'Ramirez', 'Lewis', 'Robinson', 'Walker',
    'Young', 'Allen', 'King', 'Wright', 'Scott', 'Torres', 'Nguyen', 'Hill', 'Flores',
    'Green', 'Adams', 'Nelson', 'Baker', 'Hall', 'Rivera', 'Campbell', 'Mitchell',
    'Carter', 'Roberts', 'Turner', 'Phillips', 'Evans', 'Collins', 'Stewart', 'Sanchez',
    'Morris', 'Rogers', 'Reed', 'Cook', 'Morgan', 'Bell', 'Murphy', 'Bailey',
    'Cooper', 'Richardson', 'Cox', 'Howard', 'Ward', 'Torres', 'Peterson', 'Gray',
    'Ramirez', 'James', 'Watson', 'Brooks', 'Kelly', 'Sanders', 'Price', 'Bennett',
    'Wood', 'Barnes', 'Ross', 'Henderson', 'Coleman', 'Jenkins', 'Perry', 'Powell',
    'Long', 'Patterson', 'Hughes', 'Flores', 'Washington', 'Butler', 'Simmons',
    'Foster', 'Gonzales', 'Bryant', 'Alexander', 'Russell', 'Griffin', 'Diaz',
    'Hayes', 'Myers', 'Ford', 'Hamilton', 'Graham', 'Sullivan', 'Wallace', 'Woods',
    'Cole', 'West', 'Jordan', 'Owens', 'Reynolds', 'Fisher', 'Ellis', 'Harrison',
    'Gibson', 'Mcdonald', 'Cruz', 'Marshall', 'Ortiz', 'Gomez', 'Murray', 'Freeman',
    'Wells', 'Webb', 'Simpson', 'Stevens', 'Tucker', 'Porter', 'Hunter', 'Hicks',
    'Crawford', 'Henry', 'Boyd', 'Mason', 'Morales', 'Kennedy', 'Warren', 'Dixon',
    'Ramos', 'Reyes', 'Burns', 'Gordon', 'Shaw', 'Holmes', 'Rice', 'Robertson',
    'Hunt', 'Black', 'Daniels', 'Palmer', 'Mills', 'Nichols', 'Grant', 'Knight',
    'Ferguson', 'Rose', 'Stone', 'Hawkins', 'Dunn', 'Perkins', 'Hudson', 'Spencer',
    'Gardner', 'Stephens', 'Payne', 'Pierce', 'Berry', 'Matthews', 'Arnold', 'Wagner',
    'Willis', 'Ray', 'Watkins', 'Moreno', 'Snyder', 'Hoover', 'Townsend', 'Medina',
]

TITLES = [
    'Chief Executive Officer', 'Chief Financial Officer', 'Chief Operating Officer',
    'Chief Technology Officer', 'Chief Marketing Officer', 'Chief People Officer',
    'Vice President of Operations', 'Vice President of Sales', 'Vice President of Marketing',
    'Vice President of Business Development', 'Vice President of Finance',
    'Senior Director of Strategy', 'Senior Director of Corporate Affairs',
    'Senior Director of Stakeholder Engagement', 'Senior Director of Communications',
    'Director of Public Relations', 'Director of Human Resources',
    'Director of Business Operations', 'Director of Client Success',
    'Director of Enterprise Partnerships', 'Director of Compliance',
    'Senior Manager of Corporate Initiatives', 'Senior Manager of Advocacy',
    'Senior Manager of Stakeholder Relations', 'Senior Manager of Operations',
    'Manager of Corporate Communications', 'Manager of Talent Acquisition',
    'Manager of Strategic Partnerships', 'Manager of Client Relations',
    'Senior Analyst, Corporate Strategy', 'Senior Analyst, Market Intelligence',
    'Senior Analyst, Financial Planning', 'Business Development Associate',
    'Corporate Affairs Specialist', 'Communications Specialist',
    'Operations Coordinator', 'Administrative Coordinator',
    'Executive Assistant to the CEO', 'Executive Assistant to the CFO',
    'Associate, Corporate Development', 'Associate, Public Affairs',
]

DEPARTMENTS = [
    'Executive Leadership', 'Finance & Accounting', 'Human Resources',
    'Corporate Strategy', 'Operations', 'Business Development',
    'Marketing & Communications', 'Legal & Compliance', 'Technology',
    'Client Services', 'Public Affairs', 'Administrative Services',
    'Stakeholder Relations', 'Corporate Development', 'Risk Management',
]


def generate_employee_batch(n=12):
    """Generate a batch of fake employee dicts."""
    employees = []
    # Use a time-based seed for uniqueness across visits, but pick reproducibly within the batch
    rng = random.Random()
    used_emails = set()

    for _ in range(n):
        first = rng.choice(FIRST_NAMES)
        last = rng.choice(LAST_NAMES)

        # Handle email collisions by appending a number
        base_email = f"{first.lower()}.{last.lower()}@acpwb.com"
        email = base_email
        counter = 2
        while email in used_emails:
            email = f"{first.lower()}.{last.lower()}{counter}@acpwb.com"
            counter += 1
        used_emails.add(email)

        seed = hashlib.md5(f"{first}{last}{rng.random()}".encode()).hexdigest()[:16]

        employees.append({
            'first_name': first,
            'last_name': last,
            'email': email,
            'title': rng.choice(TITLES),
            'department': rng.choice(DEPARTMENTS),
            'avatar_seed': seed,
        })

    return sorted(employees, key=lambda e: e['last_name'])
