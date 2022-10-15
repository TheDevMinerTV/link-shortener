package main

type LinkDto struct {
	ID    uint   `json:"id"`
	Short string `json:"short"`
	Long  string `json:"long"`
}

func LinkToDto(link Link) LinkDto {
	return LinkDto{
		ID:    link.ID,
		Short: link.Short,
		Long:  link.Long,
	}
}

func LinksToDto(links []Link) []LinkDto {
	dtos := make([]LinkDto, len(links))
	for i, link := range links {
		dtos[i] = LinkToDto(link)
	}

	return dtos
}
