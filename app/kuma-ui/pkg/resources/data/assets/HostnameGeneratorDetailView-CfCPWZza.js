import{d as T,r as t,o as a,m as i,w as e,b as o,e as l,p as w,R as V,s as A,c as p,F as _,v as b,U as E,t as d,q as k}from"./index-C-Llvxgw.js";import{_ as N}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-D7DuEGVs.js";const z=T({__name:"HostnameGeneratorDetailView",setup(F){return($,r)=>{const f=t("RouteTitle"),B=t("XCopyButton"),L=t("XBadge"),h=t("XLayout"),X=t("XAboutCard"),v=t("DataSource"),x=t("AppView"),D=t("DataLoader"),S=t("RouteView");return a(),i(S,{name:"hostname-generator-detail-view",params:{name:""}},{default:e(({route:g,t:c,uri:y})=>[o(f,{title:c("hostname-generators.routes.items.title"),render:!1},null,8,["title"]),r[3]||(r[3]=l()),o(D,{src:y(w(V),"/hostname-generators/:name",{name:g.params.name})},{default:e(({data:n})=>[o(x,{docs:c("hostname-generators.href.docs")},{title:e(()=>[A("h1",null,[o(B,{text:n.name},{default:e(()=>[o(f,{title:c("hostname-generators.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["text"])])]),default:e(()=>[r[2]||(r[2]=l()),o(h,{type:"stack"},{default:e(()=>[o(X,{title:c("hostname-generators.routes.item.about.title"),created:n.creationTime,modified:n.modificationTime},{default:e(()=>[(a(!0),p(_,null,b([{...n.spec.selector.meshService.matchLabels,...n.spec.selector.meshExternalService.matchLabels,...n.spec.selector.meshMultiZoneService.matchLabels}],s=>(a(),p(_,{key:typeof s},[Object.keys(s).length?(a(),i(E,{key:0,layout:"horizontal"},{title:e(()=>[l(d(c("http.api.property.tags")),1)]),body:e(()=>[o(h,{type:"separated"},{default:e(()=>[(a(!0),p(_,null,b(s,(u,m)=>(a(),i(L,{key:m},{default:e(()=>[l(d(m)+":"+d(u),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)):k("",!0)],64))),128))]),_:2},1032,["title","created","modified"]),r[1]||(r[1]=l()),o(N,{resource:n.$raw},{default:e(({copy:s,copying:u})=>[u?(a(),i(v,{key:0,src:y(w(V),"/hostname-generators/:name/as/kubernetes",{name:g.params.name},{cacheControl:"no-store"}),onChange:m=>{s(C=>C(m))},onError:m=>{s((C,R)=>R(m))}},null,8,["src","onChange","onError"])):k("",!0)]),_:2},1032,["resource"])]),_:2},1024)]),_:2},1032,["docs"])]),_:2},1032,["src"])]),_:1})}}});export{z as default};
