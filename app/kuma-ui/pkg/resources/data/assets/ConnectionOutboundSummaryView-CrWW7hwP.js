import{_ as C}from"./NavTabs.vue_vue_type_script_setup_true_lang-CJbq584y.js";import{d as R,a as t,o as l,b as m,w as e,e as o,m as b,t as p,f as c,W as y,G as D,D as O}from"./index-B_EoIyfE.js";const N=R({__name:"ConnectionOutboundSummaryView",props:{data:{},dataplaneOverview:{}},setup(u){const s=u;return(h,k)=>{const d=t("RouterLink"),_=t("DataCollection"),v=t("RouterView"),f=t("AppView"),w=t("RouteView");return l(),m(w,{name:"connection-outbound-summary-view",params:{connection:"",inactive:!1}},{default:e(({route:a,t:V})=>[o(f,null,{title:e(()=>[b("h2",null,`
          Outbound `+p(a.params.connection),1)]),default:e(()=>{var r;return[c(),o(C,{"active-route-name":(r=a.active)==null?void 0:r.name},y({_:2},[D(a.children,n=>({name:`${n.name}`,fn:e(()=>[o(d,{to:{name:n.name,query:{inactive:a.params.inactive?null:void 0}}},{default:e(()=>[c(p(V(`connections.routes.item.navigation.${n.name.split("-")[3]}`)),1)]),_:2},1032,["to"])])}))]),1032,["active-route-name"]),c(),o(v,null,{default:e(({Component:n})=>[o(_,{items:Object.entries(s.data),predicate:([i,x])=>i===a.params.connection,find:!0},{default:e(({items:i})=>[(l(),m(O(n),{data:i[0][1],"dataplane-overview":s.dataplaneOverview},null,8,["data","dataplane-overview"]))]),_:2},1032,["items","predicate"])]),_:2},1024)]}),_:2},1024)]),_:1})}}});export{N as default};
