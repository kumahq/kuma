import{d as I,k as C,V as D,a as r,o as f,b as v,w as o,e as c,m as p,f as u,t as R,l as S,B as x,s as N,v as g,a2 as B,_ as k}from"./index-36b38e0c.js";import{N as P}from"./NavTabs-f41c9db0.js";const A=t=>(N("data-v-784eee6b"),t=t(),g(),t),T={class:"summary-title-wrapper"},$=A(()=>p("img",{"aria-hidden":"true",src:B},null,-1)),j={class:"summary-title"},q=I({__name:"DataPlaneInboundSummaryView",props:{data:{}},setup(t){var l;const{t:w}=C(),b=D(),V=t,y=(((l=b.getRoutes().find(e=>e.name==="data-plane-inbound-summary-view"))==null?void 0:l.children)??[]).map(e=>{var s,a;const i=typeof e.name>"u"?(s=e.children)==null?void 0:s[0]:e,n=i.name,d=((a=i.meta)==null?void 0:a.module)??"";return{title:w(`data-planes.routes.item.navigation.${n}`),routeName:n,module:d}});return(e,i)=>{const n=r("DataCollection"),d=r("RouterView"),_=r("AppView"),s=r("RouteView");return f(),v(s,{name:"data-plane-inbound-summary-view",params:{service:""}},{default:o(({route:a})=>[c(_,null,{title:o(()=>[p("div",T,[$,u(),p("h2",j,`
            Inbound :`+R(a.params.service),1)])]),default:o(()=>[u(),c(P,{tabs:S(y)},null,8,["tabs"]),u(),c(d,null,{default:o(h=>[c(n,{items:V.data,predicate:m=>`${m.port}`===a.params.service,find:!0},{default:o(({items:m})=>[(f(),v(x(h.Component),{data:m[0]},null,8,["data"]))]),_:2},1032,["items","predicate"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const G=k(q,[["__scopeId","data-v-784eee6b"]]);export{G as default};
