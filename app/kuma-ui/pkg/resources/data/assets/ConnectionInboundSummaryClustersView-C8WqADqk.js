import{d as w,r as n,m as i,o as p,w as a,b as s,e as m,s as F,U as T,c as V,F as E,v as B}from"./index-D_WxlpfD.js";const M=w({__name:"ConnectionInboundSummaryClustersView",props:{routeName:{}},setup(l){const d=l;return(A,t)=>{const u=n("RouteTitle"),_=n("XAction"),g=n("XCodeBlock"),y=n("DataCollection"),C=n("DataLoader"),f=n("AppView"),h=n("RouteView");return p(),i(h,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,proxyType:"",mesh:"",proxy:"",connection:""},name:d.routeName},{default:a(({route:e,uri:x})=>[s(u,{render:!1,title:"Clusters"}),t[1]||(t[1]=m()),s(f,null,{default:a(()=>[s(C,{src:x(F(T),"/connections/clusters/for/:proxyType/:name/:mesh",{proxyType:{ingresses:"zone-ingress",egresses:"zone-egress"}[e.params.proxyType]??"dataplane",name:e.params.proxy,mesh:e.params.mesh||"*"})},{default:a(({data:R,refresh:k})=>[(p(!0),V(E,null,B([e.params.connection.replace("_",":")],r=>(p(),i(y,{key:typeof r,items:R.split(`
`),predicate:c=>c.startsWith(`${r}::`)},{default:a(({items:c})=>[s(g,{language:"json",code:c.map(o=>o.replace(`${r}::`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{"primary-actions":a(()=>[s(_,{action:"refresh",appearance:"primary",onClick:k},{default:a(()=>t[0]||(t[0]=[m(`
                  Refresh
                `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"]))),128))]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["name"])}}});export{M as default};
