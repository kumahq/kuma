import{d as m,L as p,f as y,o as f,g as h,w as a,h as n,i as o,af as b,aq as _,l as x,D as g}from"./index-cf0727dc.js";import{_ as k}from"./CodeBlock.vue_vue_type_style_index_0_lang-ca4abdee.js";import{t as q}from"./toYaml-4e00099e.js";const K=m({__name:"ResourceCodeBlock",props:{id:{type:String,required:!0},resource:{type:Object,required:!0},resourceFetcher:{type:Function,required:!0},codeMaxHeight:{type:String,required:!1,default:null},isSearchable:{type:Boolean,required:!1,default:!1}},setup(s){const e=s,{t:r}=p(),i=y(()=>c(e.resource));async function u(){const t=await e.resourceFetcher({format:"kubernetes"});return c(t)}function c(t){const{creationTime:l,modificationTime:T,...d}=t;return q(d)}return(t,l)=>(f(),h(k,{id:s.id,language:"yaml",code:i.value,"is-searchable":e.isSearchable,"query-key":e.id,"code-max-height":e.codeMaxHeight},{"secondary-actions":a(()=>[n(o(b),{class:"kubernetes-copy-button-tooltip",label:o(r)("common.copyKubernetesText"),placement:"bottomEnd","max-width":"200"},{default:a(()=>[n(_,{class:"kubernetes-copy-button","get-text":u,"copy-text":o(r)("common.copyKubernetesText"),"has-border":"","hide-title":""},{default:a(()=>[x(g(o(r)("common.copyKubernetesShortText")),1)]),_:1},8,["copy-text"])]),_:1},8,["label"])]),_:1},8,["id","code","is-searchable","query-key","code-max-height"]))}});export{K as _};
