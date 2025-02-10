package invalid_mergeable

type Conf struct {
    MissingOmitEmpty *string `json:"missing_omit"` // want "field Conf.MissingOmitEmpty mergeable field must have 'omitempty' in JSON tag"
    InvalidList      []struct {
        BadPtr *string `json:"bad_ptr"` // want "field Conf.InvalidList\\[\\].BadPtr mergeable field must have 'omitempty' in JSON tag"
    } `json:"invalid_list"`
}
