import{d as V,g,h as D,o as t,l as k,n as r,H as i,k as _,j as e,a6 as R,t as x,a7 as S,r as c,i as u,w as n,E as B,x as M,p as $}from"./index-23176b1b.js";const I={class:"date-status"},C=V({__name:"ResourceDateStatus",props:{creationTime:{},modificationTime:{}},setup(p){const{t:a,formatIsoDate:m}=g(),d=p,l=D(()=>m(d.creationTime)),s=D(()=>m(d.modificationTime));return(f,h)=>(t(),k("span",I,[r(i(_(a)("common.detail.created"))+": "+i(l.value)+" ",1),e(_(R)),r(" "+i(_(a)("common.detail.modified"))+": "+i(s.value),1)]))}});const N=x(C,[["__scopeId","data-v-fa366713"]]),A={key:2,class:"stack","data-testid":"detail-view-details"},E={class:"date-status-wrapper"},b=V({__name:"MeshDetailView",setup(p){const a=S();return(m,d)=>{const l=c("RouteTitle"),s=c("DataSource"),f=c("AppView"),h=c("RouteView");return t(),u(h,{name:"mesh-overview-view",params:{mesh:""}},{default:n(({route:v,t:T})=>[e(l,{title:T("meshes.routes.overview.title")},null,8,["title"]),r(),e(f,null,{default:n(()=>[e(s,{src:`/meshes/${v.params.mesh}`},{default:n(({data:o,error:w})=>[e(s,{src:`/mesh-insights/${v.params.mesh}`},{default:n(({data:y})=>[w?(t(),u(B,{key:0,error:w},null,8,["error"])):o===void 0?(t(),u(M,{key:1})):(t(),k("div",A,[e(_(a),{mesh:o,"mesh-insight":y},null,8,["mesh","mesh-insight"]),r(),$("div",E,[e(N,{"creation-time":o.creationTime,"modification-time":o.modificationTime},null,8,["creation-time","modification-time"])])]))]),_:2},1032,["src"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});const H=x(b,[["__scopeId","data-v-e0c46cc1"]]);export{H as default};
