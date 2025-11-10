## lingopaste.com

Lingopaste.com is a website which can be used to paste and share short snippets of text, with built-in LLM based translation capabilities.

It only supports sharing text by links, it does not have any social media features.

### Accounts, auth, and rate limiting

Since LLM calls are expensive, we need to limit the total requests significantly.

The goal is to have 5 pastes/day free, and then additional pastes require a paid account.

A paid account will cost $5/mo and will give you 1000 pastes/day.

We will also allow anonymous accounts, which will have IP based rate limiting.

How I envision this working technically is:

1. Anonymous account -- Identified by IP, 5 pastes / day.
2. Logged in account (google oauth login, apple oauth login, or email) -- Identified by email/oauth id/etc, 5 pastes / day.
    2a. To avoid people creating new accounts to overcome the limit, there is an additional 50 pastes/day limit on IPs that applies to all unpaid accounts, including anonymous and logged in ones.
3. Paid accounts -- Identified by logged in identifier, 1000 pastes / day.

### Translation functionality

The creator of the paste writes the paste in some language (they don't have to specify, we'll just be throwing it at an LLM API anyway), and optionally chooses a "tone" dropdown to be used for the translations. This "tone" can be "Default, Professional, Friendly, Brisque".

These options will be used as part of the LLM prompt to suggest how the AI should translate thing and imply some nuance.

The viewer of the paste will automatically see the text translated by the LLM into their preferred language, as determined by browser settings (I guess the Accept-Language header).

In addition to the paste, there will be a header that explains this is a
machine generated translation that may be inaccurate, a dropdown to select a
different language, and also a tab to toggle to the original text, or view it
side-by-side if you want.


## Technical Details

Let's talk about how this thing will actually work under the hood.

First, obviously having modern pretty HTML is key, this will be a primarily
frontend endeaver, and the frontend will make-or-break it.

I would like the pages to be responsive, which is to say choosing a new language in the dropdown to see a new translation will not load the page.

I would also like things to be very responsive, and we should assume each text snippet is fairly small (actually, let's cap it at 20k characters or so?), so it should be okay to send all translations we have cached as part of the response, and then responsively send new ones for new selected languages.

Speaking of, what do I mean by that caching bit?

Well, LLMs are expensive, so we'll ensure we only ever translate anything exactly once, and then cache it forever.

### Storage

We'll store all pastes and translations on S3.

The application will have a read-through cache for these items which has a configurable max size (say the most recently accessed 100k items, and all currently computed translations).

We'll also need one additional database for accounts, and to store metadata about pastes (such as what languages we've translated). We can also use this database as a lock when coordinating.

Let's use dynamodb for this database.

### Programming language.

I personally like Go and Rust the most, but use whatever you want.

Requirements:

1. Suitable for a responsive API for the frontend, including handling auth securely.
2. Good dynamodb and s3 API clients.
3. LLM can generate code well for the language.
4. Able to easily call LLMs.

### LLMs

We'll use the OpenAI API for this, whichever one seems best in terms of price.

We'll do the first translation on-demand, and then cache forever.

Obviously this is happening on the server.

### Payment

Please use stripe for this.

### Deployments and servers and stuff

We need to output a docker container because I'll be deploying this on K8s. I'll take care of deployment.
