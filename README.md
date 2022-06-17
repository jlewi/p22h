# Personal Knowledge Graph

An experimental project for testing some hypotheses about search and knowledge graphs.

# Hypotheses

This project is motivated by two hypotheses

1. Many search queries can be better answered by identifying and indexing entities rather than relying on full text search
1. A personal, individually managed  search/knowledge graph can be easier to adopt than a centrally managed solution

## Identifying and indexing entities

I find that a lot of my queries of personal and company knowledge stores are of the form: `Find all notes/documents mentioning X?`, where X is an entity such as a person, product, company, or system. 

Most search solutions (e.g. Google Drive Search) are based on full text search. This leads to a frustrating experience. I waste a lot of time trying different forms of my query to account for all the different aliases.

My hypothesis is that a lot of this frustration can be addressed by using named entity recognition(NER) and named entity linking(NEL) to identify mentions of an entity with higher precision and recall then just relying on full text search.

## Personal Knowledge Graphs

Today, most enterprise search engines (e.g. [Google Cloud Search](https://developers.google.com/cloud-search/docs/guides), [Elastic Workplace Search](https://www.elastic.co/workplace-search), [Glean](https://www.glean.com/), [Lucid Works](https://lucidworks.com/knowledge-management/)) appear to be selling to the enterprise not individuals. If your an employee frustrated with your search experience/knowledge management tooling you'd have to convince your CIO/CTO that this is a problem and they should invest in one of these solutions. Notably, these are not applications that a single individual can simply download and install. Some exceptions are [dala.ai](https://dala.ai/) and [CommandE](https://getcommande.com/).

I hypothesize that an individual, personally managed solution can be easier to adopt by avoiding these barriers.

# Objectives

The objectives of this project is to try to validate the above hypotheses. Concretely,

1. Can NER & NEL be used to more effectively search Google Documents?
1. Can we build solutions that can easily and cost effectively managed by an individual?

As a side goal, I'd like to learn more about frontend development and flutter.

# High Level Design

The backend is a go application that taxes care of

* Indexing documents in Google Drive
* Using [Cloud Natural Language API](https://cloud.google.com/natural-language) for entity recognition

The frontend is a flutter application providing a UI for the data.

# References

[Twitter thread asking about enterprise search](https://twitter.com/jeremylewi/status/1478708975768006659)

