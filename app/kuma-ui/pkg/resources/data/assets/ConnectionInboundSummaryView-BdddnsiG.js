import{d as C,r as o,o as l,q as m,w as n,b as s,m as b,t as d,e as r,T as v,N as x,I as N}from"./index-DxWkv34s.js";const B=C({__name:"ConnectionInboundSummaryView",props:{data:{},networking:{},routeName:{}},setup(_){const i=_;return(T,c)=>{const u=o("XAction"),w=o("XTabs"),f=o("RouterView"),y=o("AppView"),V=o("DataCollection"),g=o("RouteView");return l(),m(g,{name:i.routeName,params:{inactive:Boolean,proxyType:"",connection:""}},{default:n(({route:e,t:k})=>[s(V,{items:i.data,predicate:i.networking.type==="gateway"?t=>!0:e.params.proxyType===""?t=>`${t.name}`===e.params.connection:t=>`${t.socketAddress.replace(":","_")}`===e.params.connection,find:!0},{default:n(({items:t})=>[s(y,null,{title:n(()=>[b("h2",null,`
            Inbound `+d(e.params.connection.replace("localhost","").replace("_",":")),1)]),default:n(()=>{var p;return[c[0]||(c[0]=r()),s(w,{selected:(p=e.child())==null?void 0:p.name},v({_:2},[x(e.children,({name:a})=>({name:`${a}-tab`,fn:n(()=>[s(u,{to:{name:a,query:{inactive:e.params.inactive}}},{default:n(()=>[r(d(k(`connections.routes.item.navigation.${a.split("-")[5]}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),c[1]||(c[1]=r()),s(f,null,{default:n(a=>[(l(),m(N(a.Component),{data:t[0],networking:i.networking},null,8,["data","networking"]))]),_:2},1024)]}),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1},8,["name"])}}});export{B as default};
