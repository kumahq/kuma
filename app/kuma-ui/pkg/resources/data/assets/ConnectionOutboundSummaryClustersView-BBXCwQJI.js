import{d as w,r as n,o as p,m as i,w as a,b as t,e as m,p as F,U as T,c as V,F as E,v as B}from"./index-DIs9RbIP.js";const M=w({__name:"ConnectionOutboundSummaryClustersView",props:{routeName:{}},setup(l){const d=l;return(A,s)=>{const u=n("RouteTitle"),_=n("XAction"),g=n("XCodeBlock"),y=n("DataCollection"),C=n("DataLoader"),f=n("AppView"),h=n("RouteView");return p(),i(h,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,proxyType:"",mesh:"",proxy:"",connection:""},name:d.routeName},{default:a(({route:e,uri:x})=>[t(u,{render:!1,title:"Clusters"}),s[1]||(s[1]=m()),t(f,null,{default:a(()=>[t(C,{src:x(F(T),"/connections/clusters/for/:proxyType/:name/:mesh",{proxyType:{ingresses:"zone-ingress",egresses:"zone-egress"}[e.params.proxyType]??"dataplane",name:e.params.proxy,mesh:e.params.mesh||"*"})},{default:a(({data:R,refresh:k})=>[(p(!0),V(E,null,B([e.params.connection],r=>(p(),i(y,{key:typeof r,items:R.split(`
`),predicate:c=>c.startsWith(`${r}::`)},{default:a(({items:c})=>[t(g,{language:"json",code:c.map(o=>o.replace(`${r}::`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{"primary-actions":a(()=>[t(_,{action:"refresh",appearance:"primary",onClick:k},{default:a(()=>s[0]||(s[0]=[m(`
                  Refresh
                `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"]))),128))]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["name"])}}});export{M as default};
