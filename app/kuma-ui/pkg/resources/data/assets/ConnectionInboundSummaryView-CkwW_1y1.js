import{_ as C}from"./NavTabs.vue_vue_type_script_setup_true_lang-Cznyhy1Q.js";import{d as y,a as n,o as l,b as p,w as e,e as o,m as R,t as d,f as r,R as b,G as h,D as k}from"./index-CvRMgvyl.js";const B=y({__name:"ConnectionInboundSummaryView",props:{data:{},dataplaneOverview:{}},setup(m){const c=m;return(D,g)=>{const u=n("RouterLink"),_=n("RouterView"),v=n("AppView"),w=n("DataCollection"),f=n("RouteView");return l(),p(f,{name:"connection-inbound-summary-view",params:{connection:"",inactive:!1}},{default:e(({route:a,t:V})=>[o(w,{items:c.data,predicate:c.dataplaneOverview.dataplane.networking.type==="gateway"?i=>!0:i=>i.name===a.params.connection,find:!0},{default:e(({items:i})=>[o(v,null,{title:e(()=>[R("h2",null,`
            Inbound `+d(a.params.connection.replace("localhost","").replace("_",":")),1)]),default:e(()=>{var s;return[r(),o(C,{"active-route-name":(s=a.active)==null?void 0:s.name},b({_:2},[h(a.children,({name:t})=>({name:`${t}`,fn:e(()=>[o(u,{to:{name:t,query:{inactive:a.params.inactive?null:void 0}},"data-testid":`${t}-tab`},{default:e(()=>[r(d(V(`connections.routes.item.navigation.${t.split("-")[3]}`)),1)]),_:2},1032,["to","data-testid"])])}))]),1032,["active-route-name"]),r(),o(_,null,{default:e(t=>[(l(),p(k(t.Component),{data:i[0],"dataplane-overview":c.dataplaneOverview},null,8,["data","dataplane-overview"]))]),_:2},1024)]}),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{B as default};
