import{d as _,r as o,m as f,o as C,w as a,b as n,e as x,s as h,B as b,q as w}from"./index-D_WxlpfD.js";const k=_({__name:"ConfigurationDetailView",setup(R){return(V,r)=>{const i=o("RouteTitle"),s=o("XCodeBlock"),d=o("DataLoader"),l=o("XCard"),u=o("AppView"),p=o("RouteView");return C(),f(p,{name:"configuration-view",params:{codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:c,uri:m})=>[n(u,{breadcrumbs:[{to:{name:"configuration-view"},text:c("configuration.routes.item.breadcrumbs")}]},{title:a(()=>[w("h1",null,[n(i,{title:c("configuration.routes.item.title")},null,8,["title"])])]),default:a(()=>[r[0]||(r[0]=x()),n(l,null,{default:a(()=>[n(d,{src:m(h(b),"/config",{})},{default:a(({data:g})=>[n(s,{"data-testid":"code-block-configuration",language:"json",code:JSON.stringify(g,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:t=>e.update({codeSearch:t}),onFilterModeChange:t=>e.update({codeFilter:t}),onRegExpModeChange:t=>e.update({codeRegExp:t})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{k as default};
