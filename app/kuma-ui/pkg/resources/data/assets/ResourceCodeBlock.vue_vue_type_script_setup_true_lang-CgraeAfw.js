import{C as M}from"./CodeBlock-BVhuix4S.js";import{d as b,j as T,F as q,a5 as X,e as i,o as v,c as I,a,w as s,l as p,b as u,t as R,r as S}from"./index-Bo5vSFZC.js";const K=b({__name:"ResourceCodeBlock",props:{resource:{},codeMaxHeight:{default:void 0},isSearchable:{type:Boolean,default:!1},query:{default:""},isFilterMode:{type:Boolean,default:!1},isRegExpMode:{type:Boolean,default:!1}},emits:["query-change","filter-mode-change","reg-exp-mode-change"],setup(f,{emit:y}){const{t:c}=T(),t=f,r=y,g=q(()=>d(t.resource));function d(o){return"creationTime"in o&&delete o.creationTime,"modificationTime"in o&&delete o.modificationTime,X.stringify(o)}return(o,n)=>{const h=i("XIcon"),C=i("KCodeBlockIconButton"),x=i("XCopyButton"),B=i("XDisclosure");return v(),I("div",null,[a(M,{language:"yaml",code:g.value,"is-searchable":t.isSearchable,"code-max-height":t.codeMaxHeight,query:t.query,"is-filter-mode":t.isFilterMode,"is-reg-exp-mode":t.isRegExpMode,onQueryChange:n[0]||(n[0]=e=>r("query-change",e)),onFilterModeChange:n[1]||(n[1]=e=>r("filter-mode-change",e)),onRegExpModeChange:n[2]||(n[2]=e=>r("reg-exp-mode-change",e))},{"secondary-actions":s(()=>[a(B,null,{default:s(({expanded:e,toggle:m})=>[a(C,{"copy-tooltip":p(c)("common.copyKubernetesText"),theme:"dark",onClick:()=>{e||m()}},{default:s(()=>[a(h,{name:"copy"}),u(R(p(c)("common.copyKubernetesShortText")),1)]),_:2},1032,["copy-tooltip","onClick"]),u(),a(x,{format:"hidden"},{default:s(({copy:_})=>[S(o.$slots,"default",{copy:k=>{e&&m(),k(l=>_(d(l)),l=>console.error(l))},copying:e})]),_:2},1024)]),_:3})]),_:3},8,["code","is-searchable","code-max-height","query","is-filter-mode","is-reg-exp-mode"])])}}});export{K as _};
