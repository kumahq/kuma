import{d as C,a as n,o as r,b as p,w as e,e as o,m as O,t as m,f as i,O as h,G as y,E as x}from"./index-DPw5bDvs.js";const B=C({__name:"ConnectionOutboundSummaryView",props:{data:{},dataplaneOverview:{}},setup(d){const s=d;return(A,D)=>{const u=n("XAction"),_=n("XTabs"),v=n("DataCollection"),w=n("RouterView"),f=n("AppView"),V=n("RouteView");return r(),p(V,{name:"connection-outbound-summary-view",params:{connection:"",inactive:!1}},{default:e(({route:t,t:b})=>[o(f,null,{title:e(()=>[O("h2",null,`
          Outbound `+m(t.params.connection),1)]),default:e(()=>{var l;return[i(),o(_,{selected:(l=t.child())==null?void 0:l.name},h({_:2},[y(t.children,a=>({name:`${a.name}-tab`,fn:e(()=>[o(u,{to:{name:a.name,query:{inactive:t.params.inactive}}},{default:e(()=>[i(m(b(`connections.routes.item.navigation.${a.name.split("-")[3]}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),i(),o(w,null,{default:e(({Component:a})=>[o(v,{items:Object.entries(s.data),predicate:([c,R])=>c===t.params.connection,find:!0},{default:e(({items:c})=>[(r(),p(x(a),{data:c[0][1],"dataplane-overview":s.dataplaneOverview},null,8,["data","dataplane-overview"]))]),_:2},1032,["items","predicate"])]),_:2},1024)]}),_:2},1024)]),_:1})}}});export{B as default};
